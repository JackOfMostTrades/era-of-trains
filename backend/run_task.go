package main

import "fmt"

func runTask(server *GameServer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("need non-zero rags to runTask")
	}
	task := args[0]
	if task == "daily-summary-emails" {
		err := server.runDailyNotify()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid task: %s", task)
	}

	return nil
}
