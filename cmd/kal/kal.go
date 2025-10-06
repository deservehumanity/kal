package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type Session struct {
	StartedAt 	time.Time 	`json:"startedAt"`
	EndedAt 	time.Time 	`json:"endedAt"`
}

type Activity struct {
	Name 		string 		`json:"name"`
	Sessions 	[]Session 	`json:"sessions"`
}

type ActivityStats struct {
	Name					string		`json:"name"`
	TotalHoursSpent			float32		`json:"totalHoursSpent"`
	TotalSpentFormatted		string		`json:"totalSpentFormatted"`
	IsActive				bool		`json:"isActive"`
	ActiveSince				string		`json:"activeSince"`
}

type RunFlags struct {
	ShouldOutputJson bool
}

type Application struct {
	storage Storer
	runFlags RunFlags
	rootCmd *cobra.Command
	configDir string
}

func NewApp(storage Storer, configDir string) (*Application, error) {
	return &Application {
		storage: storage,
		runFlags: RunFlags {
			ShouldOutputJson: false,
		},
		configDir: configDir,
		rootCmd: &cobra.Command {
			Use: "kal",
			Short: "Track time spent on real-life activities. Run [kal start {activity-name}] to start tracking and [kal stop {activity-name}] to stop it.",
		},
	}, nil
}

func (a *Application) Setup() {
	newActivityCmd := &cobra.Command {
		Use: "new-activity",
		Aliases: []string{"na"},
		Short: "Creates a new activity record.",
		GroupID: "essential",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			
			activities, err := a.storage.Load()
			if err != nil {
				return err
			}
		
			newActivityName := args[0]

			for _, activity := range activities {
				if (activity.Name == newActivityName) {
					return errors.New(fmt.Sprintf("activity \"%s\" already exists", newActivityName))
				}
			}
		
			activity := Activity {
				Name: newActivityName,
				Sessions: []Session{},
			}
		
			activities = append(activities, activity)
		
			if err := a.storage.Save(activities); err != nil {
				return err
			}

			return nil
		},
	}

	startActivityCmd := &cobra.Command {
		Use: "start",
		Short: "Starts tracking given activity",
		GroupID: "essential",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing activity name")
			}

			activityName := args[0]

			activities, err := a.storage.Load()
			if err != nil {
				return err
			}
				
			var activity *Activity
		
			// Range creates a copy of activiy and then gives a copy, so use index instead of idx, activity
			for i := range activities {
				if activities[i].Name == activityName {
					activity = &activities[i]
					break
				}
			}

			if activity == nil {
				return errors.New(fmt.Sprintf("unknown activity name \"%s\"", activityName))
			}
		
			if (len(activity.Sessions) > 0 && 
				activity.Sessions[len(activity.Sessions)-1].EndedAt.IsZero()) {
				return errors.New(
					fmt.Sprintf(
						"activity \"%s\" has an unfinished session. please finish or remove it first",
						activityName,
					),
				)
			}

			activity.Sessions = append(activity.Sessions, Session {
				StartedAt: time.Now(),
			})
			
			if err := a.storage.Save(activities); err != nil {
				return err
			}

			return nil
		},
	}

	stopActivityCmd := &cobra.Command {
		Use: "stop",
		Short: "Stops tracking given activity",
		GroupID: "essential",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing activity name")
			}

			activityName := args[0]

			activities, err := a.storage.Load()
			if err != nil {
				return err
			}

			activityIndex := -1

			for idx, activity := range activities {
				if activity.Name == activityName {
					activityIndex = idx
					break;
				}
			}
		
			if (activityIndex == -1) {
				return errors.New(fmt.Sprintf("unknown activity name \"%s\"", activityName))
			}
		
			if (len(activities[activityIndex].Sessions) == 0 || 
				!activities[activityIndex].Sessions[len(activities[activityIndex].Sessions)-1].EndedAt.IsZero()) {
				return errors.New(
					fmt.Sprintf(
						"activity \"%s\" has no unfinished sessions. please create a new session",
						activityName,
					),
				)
			}

			activities[activityIndex].Sessions[len(activities[activityIndex].Sessions)-1].EndedAt = time.Now()
		
			if err := a.storage.Save(activities); err != nil {
				return err
			} 

			return nil
		},
	}
	

	removeActivityCmd := &cobra.Command {
		Use: "remove",
		Aliases: []string{"rm"},
		Short: "Removes activity record",
		GroupID: "essential",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing activity name")
			}

			activityName := args[0]
			
			activities, err := a.storage.Load()
			if err != nil {
				return err
			}

			activityIndex := -1

			for idx, activity := range activities {
				if activity.Name == activityName {
					activityIndex = idx
					break;
				}
			}
		
			if (activityIndex == -1) {
				return nil
			}
		
			activities = append(activities[:activityIndex], activities[activityIndex + 1:]...)
		
			if err := a.storage.Save(activities); err != nil {
				return err
			}

			return nil
		},
	}

	activityStatsCmd := &cobra.Command {
		Use: "stat",
		Short: "Get stats for a given activity",
		GroupID: "essential",
		RunE: func(cmd *cobra.Command, args []string) error {
			activities, err := a.storage.Load()
			if err != nil {
				return err
			}
			
			// May the universe forgive me
			if len(args) > 0 {
				activityName := args[0]
				
				var activity *Activity

				for _, a := range activities {
					if a.Name == activityName {
						activity = &a
						break;
					}
				}
		
				if activity == nil {
					return errors.New(fmt.Sprintf("unknown activity name \"%s\"", activityName))
				}
		
				stats := a.getActivityStats(activity)
				
				if a.runFlags.ShouldOutputJson {
					v, err := json.MarshalIndent(stats, "", "\t")
					if err != nil {
						return err
					}

					fmt.Println(string(v))
				} else {
					fmt.Printf("Activity: %s\n", stats.Name)
					fmt.Printf("Total time spent (formatted): %s\n", stats.TotalSpentFormatted)
					fmt.Printf("Total time spent (hours): %f\n", stats.TotalHoursSpent)
					if stats.IsActive {
						fmt.Printf("Active session since: %s\n", stats.ActiveSince)
					}
				}
			} else {
				var activitieStatsArray []ActivityStats
				
				for _, activity := range activities {
					activitieStatsArray = append(activitieStatsArray, a.getActivityStats(&activity))
				}
				
				if a.runFlags.ShouldOutputJson {
					v, err := json.MarshalIndent(activitieStatsArray, "", "\t")
					if err != nil {
						return err
					}

					fmt.Println(string(v))
				} else {
					for i, stats := range activitieStatsArray {
						fmt.Printf("Activity: %s\n", stats.Name)
						fmt.Printf("Total time spent (formatted): %s\n", stats.TotalSpentFormatted)
						fmt.Printf("Total time spent (hours): %f\n", stats.TotalHoursSpent)
						if stats.IsActive {
							fmt.Printf("Active session since: %s\n", stats.ActiveSince)
						}

						if i < len(activitieStatsArray) - 1 {
							fmt.Println()
						}
					}
				}
			}

			return nil
		},
	}

	listActivitiesCmd := &cobra.Command {
		Use: "list",
		Aliases: []string{"ls"},
		Short: "list all activity names",
		GroupID: "essential",
		RunE: func(cmd *cobra.Command, args []string) error {
			activities, err := a.storage.Load()
			if err != nil {
				return err
			}
			
			if a.runFlags.ShouldOutputJson {
				var activityNames []string
				for _, activity := range activities {
					activityNames = append(activityNames, activity.Name)
				}
				
				jsonBlob, err := json.MarshalIndent(activityNames, "", "\t") 
				if err != nil {
					return err
				}
				
				fmt.Println(string(jsonBlob))
			} else {
				for	_, activity := range activities {
					fmt.Println(activity.Name)
				}
			}

			return nil
		},
	}
	
	logActivitiesCmd := &cobra.Command {
		Use: "log",
		Short: "see all active activities",
		GroupID: "essential",
		RunE: func(cmd *cobra.Command, args []string) error {
			activities, err := a.storage.Load()
			if err != nil {
				return err
			}
			
			var unfinishedActivities []Activity

			for _, activity := range activities {
				if len(activity.Sessions) > 0 && activity.Sessions[len(activity.Sessions)-1].EndedAt.IsZero() {
					unfinishedActivities = append(unfinishedActivities, activity)
				}
			}
			
			if a.runFlags.ShouldOutputJson {
				activitiesJSON, err := json.MarshalIndent(unfinishedActivities, "", "\t")
				if err != nil {
					return err
				}

				fmt.Println(string(activitiesJSON))
			} else {
				for _, activity := range unfinishedActivities {
					fmt.Printf("Activity: %s\n", activity.Name)
					lastSession := activity.Sessions[len(activity.Sessions)-1]
					fmt.Printf("StartedAt: %s\n", lastSession.StartedAt.Format(time.RFC822))
				}
			}

			return nil
		},
	}
	
	renameActivityCmd := &cobra.Command {
		Use: "rename",
		Aliases: []string{"mv"},
		Short: "renames given activity",
		GroupID: "essential",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("expected 2 arguments. old name and new name")
			}

			oldName := args[0]
			newName := args[1]

			activities, err := a.storage.Load()
			if err != nil {
				return err
			}
			
			isFound := false

			for i := range activities {
				a := &activities[i]

				if (a.Name == oldName) {
					a.Name = newName
					isFound = true
					break
				}
			}
			
			if !isFound {
				return errors.New(fmt.Sprintf("no activity with name \"%s\" found", oldName))
			}

			if err := a.storage.Save(activities); err != nil {
				return err
			}

			return nil
		},
	}

	activityStatsCmd.Flags().BoolVar(
		&a.runFlags.ShouldOutputJson,
		"json",
		false,
		"output as json to stdout",
	)
	
	logActivitiesCmd.Flags().BoolVar(
		&a.runFlags.ShouldOutputJson,
		"json",
		false,
		"output as json to stdout",
	)

	listActivitiesCmd.Flags().BoolVar(
		&a.runFlags.ShouldOutputJson,
		"json",
		false,
		"output as json to stdout",
	)

	a.rootCmd.AddGroup(&cobra.Group {
		Title: "essential",
		ID: "essential",
	})

	a.rootCmd.AddCommand(newActivityCmd)
	a.rootCmd.AddCommand(startActivityCmd)
	a.rootCmd.AddCommand(stopActivityCmd)
	a.rootCmd.AddCommand(removeActivityCmd)
	a.rootCmd.AddCommand(activityStatsCmd)
	a.rootCmd.AddCommand(listActivitiesCmd)
	a.rootCmd.AddCommand(logActivitiesCmd)
	a.rootCmd.AddCommand(renameActivityCmd)
}

func (a *Application) Execute() error {
	return a.rootCmd.Execute()
}

func (a *Application) getActivityStats(activity *Activity) ActivityStats {
	var totalDuration time.Duration
	var hasActiveSession bool
	var activeSince time.Time
	for _, session := range activity.Sessions {
		if session.EndedAt.IsZero() {
			hasActiveSession = true
			activeSince = session.StartedAt
			totalDuration += time.Since(session.StartedAt)
		} else {
			totalDuration += session.EndedAt.Sub(session.StartedAt)
		}
	}

	return ActivityStats {
		Name: activity.Name,
		TotalHoursSpent: float32(totalDuration.Hours()),
		TotalSpentFormatted: formatDuration(totalDuration),
		IsActive: hasActiveSession,
		ActiveSince: activeSince.Format(time.RFC822),
	}
}

func formatDuration(d time.Duration) string {
	h := int (d.Hours())
	m := int (d.Minutes()) % 60
	s := int (d.Seconds()) % 60

	return fmt.Sprintf("%02dh %02dm %02ds", h, m ,s)
}

func formatDurationHMS(d time.Duration) (int, int, int) {
	h := int (d.Hours())
	m := int (d.Minutes()) % 60
	s := int (d.Seconds()) % 60

	return h, m, s
}

func main() {
	// TODO: Use SQLite for rich queries (make a storer interface?)
	// TODO: Automate testing
	// TODO: Maybe test with swaybar in hyprland via hooks
	// TODO: Add custom hooks for users to define
	
	hd, err := os.UserHomeDir()
	if err != nil {
		return
	}
	
	configDir := fmt.Sprintf("%s/%s", hd, ".kal")
	
	if _, err := os.Stat(configDir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(configDir, os.ModePerm); err != nil {
				return 			
			}
		} else {
			return 	
		}
	}

	storage, err := NewJSONStorage(configDir)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	app, err := NewApp(storage, configDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	app.Setup()
	app.Execute()
}
