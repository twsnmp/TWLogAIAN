package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gravwell/gravwell/v3/client"
	"github.com/gravwell/gravwell/v3/client/objlog"
	"github.com/gravwell/gravwell/v3/client/types"
)

func (b *App) readLogFromGravwell(lf *LogFile) error {
	a := strings.SplitN(lf.LogSrc.Server, "://", 2)
	if len(a) < 2 {
		return fmt.Errorf("server format error. serrve=")
	}
	sv := a[1]
	c, err := client.NewClient(sv, false, strings.Contains(a[1], "https"), &objlog.NilObjLogger{})
	if err != nil {
		return err
	}
	defer c.Close()

	// Log in
	if err = c.Login(lf.LogSrc.User, lf.LogSrc.Password); err != nil {
		return err
	}

	// Call Sync to update client internal data
	if err = c.Sync(); err != nil {
		return err
	}
	// Parse the search and make sure they're using one of the basic renderers
	psr, err := c.ParseSearchWithResponse(lf.LogSrc.Pattern, []types.FilterRequest{})
	if err != nil {
		return err
	}
	if psr.RenderModule != "text" {
		return fmt.Errorf("not supported format")
	}
	end := time.Now()
	start := end.Add(time.Hour * -1)
	if lf.LogSrc.Start != "" {
		if t, err := time.Parse("2006-01-02T15:04 MST", lf.LogSrc.Start+" JST"); err != nil {
			start = t
		}
	}
	if lf.LogSrc.End != "" {
		if t, err := time.Parse("2006-01-02T15:04 MST", lf.LogSrc.End+" JST"); err != nil {
			end = t
		}
	}
	// Now start the search
	s, err := c.StartSearch(lf.LogSrc.Pattern, start, end, false)
	if err != nil {
		return err
	}

	// Wait for the search to be completed
	if err = c.WaitForSearch(s); err != nil {
		return err
	}
	results, err := c.GetTextResults(s, 0, uint64(1000000))
	if err != nil {
		return err
	}
	logs := []string{}
	for _, r := range results.Entries {
		logs = append(logs, string(r.Data))
	}
	b.readOneLogFile(lf, strings.NewReader(strings.Join(logs, "\n")))
	return nil
}
