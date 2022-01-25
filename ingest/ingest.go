package ingest

import (
	"context"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	spreadscore "github.com/paralin/replay-spreadsheet"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Ingest monitors the spreadsheet and ingests submissions to dota replay.
type Ingest struct {
	le               *logrus.Entry
	sheet            *spreadscore.Spreadsheet
	submissionsTable *mgo.Collection
}

// NewIngest builds an Ingest instance
func NewIngest(
	le *logrus.Entry,
	sheet *spreadscore.Spreadsheet,
	submissionColl *mgo.Collection,
) *Ingest {
	return &Ingest{
		le:               le,
		sheet:            sheet,
		submissionsTable: submissionColl,
	}
}

// Execute executes the ingest instance.
func (i *Ingest) Execute(ctx context.Context) error {
	syncPeriod := time.Second * 10
	syncTicker := time.NewTicker(syncPeriod)
	defer syncTicker.Stop()

	var startIndex int
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-syncTicker.C:
		}

		// Check for new submissions
		subs, err := i.sheet.FetchSubmissions(ctx, startIndex)
		if err != nil {
			i.le.WithError(err).Warn("could not fetch submissions")
			continue
		}
		if len(subs) != 0 {
			if err := i.processSubmissions(ctx, startIndex, subs); err != nil {
				i.le.WithError(err).Warn("could not process submissions, will retry")
				continue
			} else {
				startIndex += len(subs)
			}
		}
	}
}

// processSubmissions processes incoming submissions.
func (i *Ingest) processSubmissions(ctx context.Context, baseIdx int, subs []*spreadscore.Submission) error {
	var indexes []int
	for si, sub := range subs {
		subIdx := baseIdx + si
		if sub.Imported {
			continue
		}
		sub.ShowName = GetIDFromShowName(sub.ShowName)
		le := i.le.WithField("match-id", sub.MatchID)
		if err := sub.Validate(); err != nil {
			le.WithError(err).Warn("ignoring invalid submission")
		} else {
			if err := i.writeEntryToDb(sub); err != nil {
				if mgo.IsDup(err) {
					le.Debug("entry already existed in database")
				} else {
					le.WithError(err).Warn("unable to write entry to database")
					continue
				}
			}
		}
		indexes = append(indexes, subIdx)
		le.Debug("will update row to imported=true")
	}
	if len(indexes) != 0 {
		if err := i.sheet.SetSubmissionsImported(indexes, true); err != nil {
			return errors.Wrap(err, "set submissions as imported")
		}
		i.le.Debugf("batch updated %d rows to set imported=true", len(indexes))
	}
	return nil
}

// writeEntryToDb writes an entry to the database if it doesn't exist.
func (i *Ingest) writeEntryToDb(sub *spreadscore.Submission) error {
	n, err := i.submissionsTable.Find(bson.M{"matchid": sub.MatchID}).Count()
	if err != nil && err != mgo.ErrNotFound {
		return err
	}
	if n != 0 {
		return nil
	}
	if sub.Id == "" {
		sub.Id = randString(8)
	}
	err = i.submissionsTable.Insert(sub)
	return err
}
