package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/globalsign/mgo"
	scoring "github.com/paralin/replay-spreadsheet"
	"github.com/paralin/replay-spreadsheet/gspread"
	"github.com/paralin/replay-spreadsheet/ingest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var listenPort int
var spreadsheetID = "1ez44IKV1pHBzY9axSg_ysyUXsN6ZMLc4znZT_AXAr18"
var mongoURL string

var (
	log = logrus.New()
)

func main() {
	log.SetLevel(logrus.DebugLevel)
	app := cli.NewApp()
	app.Name = "apex-t1-spreadsheet"
	app.Usage = "convert t1 apex event spreadsheet into json"
	app.Authors = []cli.Author{
		cli.Author{Name: "Christian Stewart", Email: "christian@paral.in"},
	}
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "spreadsheet-id",
			Usage:       "set the google spreadsheet id",
			Destination: &spreadsheetID,
			Value:       spreadsheetID,
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "daemon",
			Usage:  "run the watching daemon",
			Action: runDaemon,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "mongo-url",
					Usage:       "the mongo URL for database",
					EnvVar:      "MONGO_URL",
					Destination: &mongoURL,
				},
			},
		},
		cli.Command{
			Name:   "dump",
			Usage:  "dump the data once to the shell",
			Action: runDump,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// loadSheet loads the spreadsheet.
func loadSheet(ctx context.Context) (*scoring.Spreadsheet, error) {
	service, err := gspread.LoadOrAuthClient(ctx, "credentials.json")
	if err != nil {
		return nil, err
	}

	// https://docs.google.com/spreadsheets/d/:ID_IS_HERE/edit
	return scoring.NewSpreadsheet(service, spreadsheetID), nil
}

func runDaemon(c *cli.Context) error {
	le := logrus.NewEntry(log)
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	sheet, err := loadSheet(ctx)
	if err != nil {
		return err
	}
	le.Debugf("loaded spreadsheet %s", spreadsheetID)

	le.Debug("dialing mongo")
	sess, err := mgo.Dial(mongoURL)
	if err != nil {
		return errors.Wrap(err, "dial mongo")
	}
	le.Debug("dialed mongo")
	ig := ingest.NewIngest(le, sheet, sess.DB("").C("submissions"))
	return ig.Execute(ctx)
}

func runDump(c *cli.Context) error {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	sheet, err := loadSheet(ctx)
	if err != nil {
		return err
	}

	submissions, err := sheet.FetchSubmissions(ctx, 0)
	if err != nil {
		return err
	}

	jdata, err := json.MarshalIndent(submissions, "", "\t")
	if err != nil {
		return err
	}
	fmt.Println(string(jdata))
	return nil
}
