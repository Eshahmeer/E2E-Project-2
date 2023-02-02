// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package caldav

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"gopkg.in/d4l3k/messagediff.v1"
)

func TestParseTaskFromVTODO(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name      string
		args      args
		wantVTask *models.Task
		wantErr   bool
	}{
		{
			name: "normal",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Todo #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
			},
		},
		{
			name: "With priority",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
PRIORITY:9
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Todo #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Priority:    1,
				Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
			},
		},
		{
			name: "With categories",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
CATEGORIES:cat1,cat2
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Todo #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Labels: []*models.Label{
					{
						Title: "cat1",
					},
					{
						Title: "cat2",
					},
				},
				Updated: time.Unix(1543626724, 0).In(config.GetTimeZone()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTaskFromVTODO(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTaskFromVTODO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff, equal := messagediff.PrettyDiff(got, tt.wantVTask); !equal {
				t.Errorf("ParseTaskFromVTODO() gotVTask = %v, want %v, diff = %s", got, tt.wantVTask, diff)
			}
		})
	}
}

func TestGetCaldavTodosForTasks(t *testing.T) {
	type args struct {
		list  *models.ProjectWithTasksAndBuckets
		tasks []*models.TaskWithComments
	}
	tests := []struct {
		name       string
		args       args
		wantCaldav string
	}{
		{
			name: "Format single Task as Caldav",
			args: args{
				list: &models.ProjectWithTasksAndBuckets{
					Project: models.Project{
						Title: "List title",
					},
				},
				tasks: []*models.TaskWithComments{
					{
						Task: models.Task{
							Title:       "Task 1",
							UID:         "randomuid",
							Description: "Description",
							Priority:    3,
							Created:     time.Unix(1543626721, 0).In(config.GetTimeZone()),
							DueDate:     time.Unix(1543626722, 0).In(config.GetTimeZone()),
							StartDate:   time.Unix(1543626723, 0).In(config.GetTimeZone()),
							EndDate:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
							Updated:     time.Unix(1543626725, 0).In(config.GetTimeZone()),
							DoneAt:      time.Unix(1543626726, 0).In(config.GetTimeZone()),
							RepeatAfter: 86400,
							Labels: []*models.Label{
								{
									ID:    1,
									Title: "label1",
								},
								{
									ID:    2,
									Title: "label2",
								},
							},
						},
					},
				},
			},
			wantCaldav: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:List title
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011205Z
SUMMARY:Task 1
DTSTART:20181201T011203Z
DTEND:20181201T011204Z
DESCRIPTION:Description
COMPLETED:20181201T011206Z
STATUS:COMPLETED
DUE:20181201T011202Z
CREATED:20181201T011201Z
PRIORITY:3
RRULE:FREQ=SECONDLY;INTERVAL=86400
CATEGORIES:label1,label2
LAST-MODIFIED:20181201T011205Z
END:VTODO
END:VCALENDAR`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCaldavTodosForTasks(tt.args.list, tt.args.tasks)
			if diff, equal := messagediff.PrettyDiff(got, tt.wantCaldav); !equal {
				t.Errorf("GetCaldavTodosForTasks() gotVTask = %v, want %v, diff = %s", got, tt.wantCaldav, diff)
			}
		})
	}
}
