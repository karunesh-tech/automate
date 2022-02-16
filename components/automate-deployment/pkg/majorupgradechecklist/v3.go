package majorupgradechecklist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/chef/automate/components/automate-cli/pkg/status"
	"github.com/chef/automate/components/automate-deployment/pkg/cli"
	"github.com/pkg/errors"
)

const (
	initMsg = `This is a Major upgrade. 
	In this release Embedded PostgreSQL is upgraded to version 13.5 
	Which will need special care if you use Embedded PostgreSQL. 
	*Your installation is using Embedded PostgreSQL.* 
	Please confirm this check list that you have completed these steps before continuing with the Upgrade to version %s:`

	downTimeError = "There will be a downtime while upgrading. Please preare for down time and run the upgrade"

	backupError = "Please take a backup and restart the upgrade process."

	run_chef_automate_upgrade_status = `Check the status of your upgrade using:  
		   $ chef-automate upgrade status 
	
		 This should return: Automate is up-to-date`

	run_pg_data_migrate = `Migrate Data from PG 9.6 to PG 13.5 using this command: 
		   $ ` + run_pg_data_migrate_cmd

	run_pg_data_migrate_cmd = `chef-automate post-major-upgrade migrate --data=pg`

	run_chef_automate_status = `Check all services are running using: 
		   $ chef-automate status`

	check_ui = `Check Automate UI everything is running and all data is visible.`

	run_pg_data_cleanup = `If you are sure all data is available in Upgraded Automate, then we can free up old PostgreSQL 9.6 Data by running: 
		   $ ` + run_pg_data_cleanup_cmd

	run_pg_data_cleanup_cmd        = `chef-automate post-major-upgrade clear-data --pg-data`
	v3_post_checklist_confirmation = `*In case of any errors, please refer to docs.chef.io and release notes for this version.*
	Now, upgrade will start, Please confirm to continue...`
)

var postChecklist = []PostCheckList{
	{
		Id:         "migrate_pg",
		Msg:        run_pg_data_migrate,
		Cmd:        run_pg_data_migrate_cmd,
		IsExecuted: false,
	}, {
		Id:         "clean_up",
		Msg:        run_pg_data_cleanup,
		Cmd:        run_pg_data_cleanup_cmd,
		Optional:   true,
		IsExecuted: false,
	},
}

type V3ChecklistManager struct {
	writer  cli.FormatWriter
	version string
}

func NewV3ChecklistManager(writer cli.FormatWriter, version string) *V3ChecklistManager {
	return &V3ChecklistManager{
		writer:  writer,
		version: version,
	}
}
func prechecklist() []Checklist {
	return []Checklist{
		downTimeCheck(),
		backupCheck(),
	}
}

func (ci *V3ChecklistManager) RunChecklist() error {

	checklists := []Checklist{}
	checklists = append(checklists, prechecklist()...)
	checklists = append(checklists, ci.showPostChecklist(), promptUpgradeContinue())

	helper := ChecklistHelper{
		Writer: ci.writer,
	}

	ci.writer.Println(fmt.Sprintf(initMsg, ci.version)) //display the init message

	for _, item := range checklists {
		if item.TestFunc == nil {
			continue
		}
		if err := item.TestFunc(helper); err != nil {
			return errors.Wrap(err, "one of the checklist was not accepted/satisfied for upgrade")
		}
	}
	return nil
}

func (ci *V3ChecklistManager) showPostChecklist() Checklist {
	return Checklist{
		Name:        "Show_Post_Checklist",
		Description: "Show Post Checklist",
		TestFunc: func(h ChecklistHelper) error {
			for _, item := range postChecklist {
				if !item.IsExecuted {
					h.Writer.Println(item.Msg)
				}
			}
			return nil
		},
	}

}

func downTimeCheck() Checklist {
	return Checklist{
		Name:        "down_time_acceptance",
		Description: "confirmation for downtime",
		TestFunc: func(h ChecklistHelper) error {
			resp, err := h.Writer.Confirm("You had planned for a downtime?:")
			if err != nil {
				h.Writer.Error(err.Error())
				return status.Errorf(status.InvalidCommandArgsError, err.Error())
			}
			if !resp {
				h.Writer.Error(downTimeError)
				return status.New(status.InvalidCommandArgsError, downTimeError)
			}
			return nil
		},
	}
}

func backupCheck() Checklist {
	return Checklist{
		Name:        "backup_acceptance",
		Description: "confirmation check for creating a backup",
		TestFunc: func(h ChecklistHelper) error {
			resp, err := h.Writer.Confirm("You have taken backup of your data and is safe, preferred on other disk or location?")
			if err != nil {
				h.Writer.Error(err.Error())
				return status.Errorf(status.InvalidCommandArgsError, err.Error())
			}
			if !resp {
				h.Writer.Error(backupError)
				return status.New(status.InvalidCommandArgsError, downTimeError)
			}
			return nil
		},
	}
}

func promptUpgradeContinue() Checklist {
	return Checklist{
		Name:        "post_checklist",
		Description: "display post checklist and ask for final confirmation",
		TestFunc: func(h ChecklistHelper) error {
			resp, err := h.Writer.Confirm(v3_post_checklist_confirmation)
			if err != nil {
				h.Writer.Error(err.Error())
				return status.Errorf(status.InvalidCommandArgsError, err.Error())
			}
			if !resp {
				h.Writer.Error("end user not ready to upgrade")
				return status.New(status.InvalidCommandArgsError, "end user not ready to upgrade")
			}
			return nil
		},
	}
}

func (ci *V3ChecklistManager) CreatePostChecklistFile() error {
	params := PerPostChecklist{}
	params.PostChecklist = append(params.PostChecklist, postChecklist...)
	var buffer bytes.Buffer
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	buffer.Write(data)
	buffer.WriteString("\n")
	err = ioutil.WriteFile("/hab/svc/deployment-service/var/upgrade_metadata.json", buffer.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}
func (ci *V3ChecklistManager) ReadPostChecklistById(id string) (bool, error) {
	ChecklistId_Found := false
	res, err := ReadJsonFile()
	if err != nil {
		return false, err
	}
	for i := 0; i < len(res.PostChecklist); i++ {

		if res.PostChecklist[i].Id == id {
			ChecklistId_Found = res.PostChecklist[i].IsExecuted
			break

		}
	}
	return ChecklistId_Found, nil
}

func ReadJsonFile() (*PerPostChecklist, error) {
	byteValue, err := ioutil.ReadFile("/hab/svc/deployment-service/var/upgrade_metadata.json")
	if err != nil {
		return nil, err
	}
	params := PerPostChecklist{}

	err = json.Unmarshal(byteValue, &params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func (ci *V3ChecklistManager) ReadPostChecklistFile() ([]string, error) {
	var postCmdList []string
	var showPostChecklist = false
	res, err := ReadJsonFile()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(res.PostChecklist); i++ {
		if !res.PostChecklist[i].Optional && !res.PostChecklist[i].IsExecuted {
			showPostChecklist = true
			break
		}
	}

	if showPostChecklist == true {
		for i := 0; i < len(res.PostChecklist); i++ {
			if !res.PostChecklist[i].IsExecuted {
				postCmdList = append(postCmdList, res.PostChecklist[i].Msg)
			}
		}
	}
	return postCmdList, nil
}

func (ci *V3ChecklistManager) UpdatePostChecklistFile(id string) error {
	byteValue, _ := ioutil.ReadFile("test.json")
	params := PerPostChecklist{}

	json.Unmarshal(byteValue, &params)

	for i, v := range params.PostChecklist {
		if v.Id == id {
			params.PostChecklist[i].IsExecuted = true
		}
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("test.json", data, 0644)
	if err != nil {
		return err
	}
	return nil
}