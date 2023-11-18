package main

type (
	tAction int
)

const (
	appActionUnknown tAction = iota
	appActionUpgrade
	appActionAdd
	appActionDeleteRollback
	appActionRollback
	appActionResumeUpgrade

	appActionMax // all appAction enumerations should be added before this
)

func (action tAction) isValidAction() (result bool) {
	result = action > appActionUnknown && action < appActionMax
	return
}

func (action tAction) String() (result string) {
	result = []string{"Unknown", "Upgrade", "Add", "Delete", "Rollback", "Resume", "Max"}[action]
	return
}
