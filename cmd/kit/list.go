package main

import (
	"fmt"

	"github.com/dvyukov/kit/pkg/journal"
	_ "github.com/dvyukov/kit/pkg/kitp"
)

func init() {
	register("list", listRun, nil, nil, "list journal entries")
}

func listRun() error {
	journal, err := journal.Open(workdir)
	if err != nil {
		return fmt.Errorf("failed to open journal: %v", err)
	}
	for _, msg := range journal.Messages() {
		fmt.Printf("%v\n", msg)
	}
	return nil
}
