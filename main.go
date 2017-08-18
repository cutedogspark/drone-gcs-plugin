package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version = "1.0.0" // build number set at compile-time
)

func main() {
	app := cli.NewApp()
	app.Name = "google cloud storage plugin"
	app.Usage = "google cloud storage plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "gcs-key",
			Usage:  "google cloud storage credentials file",
			EnvVar: "GCS_KEY,PLUGIN_KEY",
		},
		cli.StringFlag{
			Name:   "bucket",
			Usage:  "google cloud storage bucket name",
			EnvVar: "GCS_BUCKET,PLUGIN_BUCKET",
		},
		cli.StringFlag{
			Name:   "source",
			Usage:  "upload files from source folder",
			EnvVar: "GCS_SOURCE,PLUGIN_SOURCE",
		},
		cli.StringFlag{
			Name:   "strip-prefix",
			Usage:  "strip the prefix from the target",
			EnvVar: "GCS_STRIP_PREFIX,PLUGIN_STRIP_PREFIX",
		},
		cli.StringFlag{
			Name:   "target",
			Usage:  "upload files to target folder",
			EnvVar: "GCS_TARGET,PLUGIN_TARGET",
		},
		cli.BoolFlag{
			Name:   "target-auto-date",
			Usage:  "target folder auto create current date folder(global setting)",
			EnvVar: "GCS_TARGET_DATE_FOLDER,PLUGIN_TARGET_DATE_FOLDER",
		},

		cli.StringFlag{
			Name:   "trigger-branch",
			Usage:  "trigger branch from submodule",
			EnvVar: "GCS_TRIGGER_BRANCH,PLUGIN_TRIGGER_BRANCH",
		},
		cli.StringFlag{
			Name:   "trigger-folder",
			Usage:  "trigger save folder",
			EnvVar: "GCS_TRIGGER_FOLDER,PLUGIN_TRIGGER_FOLDER",
		},

		cli.StringFlag{
			Name:   "tag-module",
			Usage:  "tag module from submodule",
			EnvVar: "GCS_TAG_MODULE,PLUGIN_TAG_MODULE",
		},
		cli.StringFlag{
			Name:   "tag-folder",
			Usage:  "tag save folder",
			EnvVar: "GCS_TAG_FOLDER,PLUGIN_TAG_FOLDER",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "show debug",
			EnvVar: "GCS_DEBUG,PLUGIN_DEBUG",
		},

		cli.StringFlag{
			Name:   "repo.owner",
			Usage:  "repository owner",
			EnvVar: "DRONE_REPO_OWNER",
		},
		cli.StringFlag{
			Name:   "repo.name",
			Usage:  "repository name",
			EnvVar: "DRONE_REPO_NAME",
		},

		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
			Value:  "unsetSHA",
		},
		cli.StringFlag{
			Name:   "commit.ref",
			Value:  "refs/heads/master",
			Usage:  "git commit ref",
			EnvVar: "DRONE_COMMIT_REF",
		},
		cli.StringFlag{
			Name:   "commit.branch",
			Value:  "master",
			Usage:  "git commit branch",
			EnvVar: "DRONE_COMMIT_BRANCH",
		},
		cli.StringFlag{
			Name:   "commit.author",
			Usage:  "git author name",
			EnvVar: "DRONE_COMMIT_AUTHOR",
			Value:  "unknown author",
		},
		cli.StringFlag{
			Name:   "commit.message",
			Usage:  "commit message",
			EnvVar: "DRONE_COMMIT_MESSAGE",
			Value:  "unset message",
		},
		cli.StringFlag{
			Name:   "build.event",
			Value:  "push",
			Usage:  "build event",
			EnvVar: "DRONE_BUILD_EVENT",
		},
		cli.IntFlag{
			Name:   "build.number",
			Usage:  "build number",
			EnvVar: "DRONE_BUILD_NUMBER",
		},
		cli.StringFlag{
			Name:   "build.status",
			Usage:  "build status",
			Value:  "success",
			EnvVar: "DRONE_BUILD_STATUS",
		},
		cli.StringFlag{
			Name:   "build.link",
			Usage:  "build link",
			EnvVar: "DRONE_BUILD_LINK",
		},
		cli.Int64Flag{
			Name:   "build.started",
			Usage:  "build started",
			EnvVar: "DRONE_BUILD_STARTED",
		},
		cli.Int64Flag{
			Name:   "build.created",
			Usage:  "build created",
			EnvVar: "DRONE_BUILD_CREATED",
		},
		cli.StringFlag{
			Name:   "build.tag",
			Usage:  "build tag",
			EnvVar: "DRONE_TAG",
		},

		cli.Int64Flag{
			Name:   "job.started",
			Usage:  "job started",
			EnvVar: "DRONE_JOB_STARTED",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	plugin := Plugin{}

	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	plugin = Plugin{
		Repo: Repo{
			Owner: c.String("repo.owner"),
			Name:  c.String("repo.name"),
		},
		Build: Build{
			Tag:     c.String("build.tag"),
			Number:  c.Int("build.number"),
			Event:   c.String("build.event"),
			Status:  c.String("build.status"),
			Commit:  c.String("commit.sha"),
			Ref:     c.String("commit.ref"),
			Branch:  c.String("commit.branch"),
			Author:  c.String("commit.author"),
			Message: c.String("commit.message"),
			Link:    c.String("build.link"),
			Started: c.Int64("build.started"),
			Created: c.Int64("build.created"),
		},
		Job: Job{
			Started: c.Int64("job.started"),
		},

		Credentials:      c.String("gcs-key"),
		Bucket:           c.String("bucket"),
		Source:           c.String("source"),
		StripPrefix:      c.String("strip-prefix"),
		Target:           c.String("target"),
		TargetDateFolder: c.Bool("target-auto-date"),

		TriggerFolder: c.String("trigger-folder"),
		TagFolder:     c.String("tag-folder"),

		// read from environmental variables
		TriggerModule: os.Getenv("T_MODULE"),
		TriggerEven:   os.Getenv("T_EVEN"),
		TriggerBranch: os.Getenv("T_BRANCH"),
		TriggerSHA:    os.Getenv("T_SHA"),

		Access:  c.StringSlice("acl"),
		Exclude: c.StringSlice("exclude"),

		Compress: c.StringSlice("compress"),
	}

	log.WithFields(log.Fields{
		"bucket":           plugin.Bucket,
		"source":           plugin.Source,
		"target":           plugin.Target,
		"triggerFolder":    plugin.TriggerFolder,
		"tagFolder":        plugin.TagFolder,
		"targetDateFolder": plugin.TargetDateFolder,
		"triggerModule":    plugin.TriggerModule,
		"triggerEven":      plugin.TriggerEven,
		"triggerBranch":    plugin.TriggerBranch,
		"triggerSHA":       plugin.TriggerSHA,
		"buildEvent":       plugin.Build.Event,
	}).Debug("Parameter..")

	return plugin.Exec()
}
