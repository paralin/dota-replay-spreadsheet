package spreadscore

import "time"

import "github.com/pkg/errors"

// EarliestTime is the earliest acceptable timestamp.
var EarliestTime = time.Date(2019, 1, 1, 1, 1, 1, 0, time.UTC)

// EarliestMatchID is the earliest acceptable match ID.
var EarliestMatchID int64 = 5118704000

// Submission contains information about a submission.
type Submission struct {
	// Id is the identifier field.
	Id string `json:"_id" bson:"_id"`
	// Timestamp is the timestamp of the submission.
	Timestamp time.Time `json:"timestamp" bson:"createdAt"`
	// MatchID is the match ID of the match.
	MatchID int64 `json:"matchID" bson:"matchid"`
	// Name is the nickname of the submitter.
	Name string `json:"name" bson:"uname"`
	// Description is the description by the submitter.
	Description string `json:"description" bson:"description"`
	// ShowName is the show name specified by the submitter.
	// Should be transformed into the correct ID before submission to mongo.
	ShowName string `json:"showName" bson:"show"`
	// Hero is the hero to watch specified by the submitter.
	Hero string `json:"hero" bson:"hero_to_watch"`
	// Imported indicates the entry was imported to the replay db.
	Imported bool `json:"imported" bson:"-"`
	// Reviewed indicates if the submission has been reviewed yet.
	Reviewed bool `json:"reviewed" bson:"reviewed"`
	// Status is the status of the replay.
	Status int `json:"status" bson:"status"`
	// MatchTime is the time the event occurs in the match in seconds.
	MatchTime int `json:"matchtime" bson:"matchtime"`
	// IngameTime is the time the event occurs in-game, same as matchtime but a string.
	IngameTime string `json:"ingame_time" bson:"ingame_time"`
}

// Validate validates the submission
func (s *Submission) Validate() error {
	if s.Timestamp.Before(EarliestTime) {
		return errors.Errorf(
			"timestamp too early: %s < %s",
			s.Timestamp.String(),
			EarliestTime.String(),
		)
	}
	if s.MatchID < EarliestMatchID {
		return errors.Errorf(
			"match id too low: %d < %d",
			s.MatchID,
			EarliestMatchID,
		)
	}
	return nil
}
