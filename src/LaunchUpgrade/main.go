package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"softwareupgrade"
	"strings"
	"time"
)

var (
	commonSSLcertContent                                     []byte
	userSSLcertContent                                       []byte
	appStatus                                                string
	debugLogFilename, failedNodesFilename                    string
	rollbackInfoFilename                                     string
	jsonFilename                                             string
	debug                                                    bool
	disableNodeVerification, disableFileVerification, dryRun bool
	disableTargetDirVerification                             bool
	mode, rollbackSuffix                                     string
	action                                                   tAction
)

func upgradeOrRollback(jsonContents []byte) {
	var upgradeconfig softwareupgrade.UpgradeConfig
	// Parse the JSON
	json.Unmarshal(jsonContents, &upgradeconfig)

	if upgradeconfig.Common.SSHTimeout != "" {
		parsedTimeout, err := time.ParseDuration(upgradeconfig.Common.SSHTimeout)
		if err == nil {
			softwareupgrade.SetSSHTimeout(parsedTimeout)
		} else {
			softwareupgrade.SetSSHTimeout(5 * time.Second)
		}
	} else {
		softwareupgrade.SetSSHTimeout(5 * time.Second)
	}

	DebugLog.Println("This session PID: %d rollback file: %s", os.Getpid(), rollbackInfoFilename)

	if !disableFileVerification {
		if err := upgradeconfig.VerifyFilesExist(); err != nil {
			DebugLog.Printf("%v\n", err)
			return
		}
		DebugLog.Println("All source files verified.")
	}

	// GroupNames is the name given to each combination of software
	SoftwareGroupNames := upgradeconfig.GetGroupNames()
	DebugLog.Println("%d groups defined: %v", len(SoftwareGroupNames), SoftwareGroupNames)

	// Nodes contains the list of the nodes to upgrade.
	nodes := upgradeconfig.GetNodes()
	DebugLog.Println("%d nodes found: %v", len(nodes), nodes)

	if !disableNodeVerification {
		// Verify all nodes can be looked up using IP address.
		var (
			msg                  string
			failCount, nodeCount int
		)
		for _, node := range nodes {
			_, err := net.LookupIP(node)
			if err != nil {
				msg = fmt.Sprintf("%sCan't resolve %s\n", msg, node)
				failCount++
			} else {
				nodeCount++
			}
		}
		if msg != "" {
			DebugLog.Print(msg)
			return
		}
		if nodeCount == len(nodes) && nodeCount > 0 && failCount == 0 {
			DebugLog.Println("All nodes verified to be resolvable to IP addresses.")
		}
	}

	if !disableTargetDirVerification || action == appActionAdd {
		// Only perform directory verification if there is at least 1 node
		if nodeCount := upgradeconfig.GetNodeCount(); nodeCount > 0 {
			DebugLog.Println("Verifying target directories, please wait.")

			// Verify all target directories exist. This is also an opportunity
			// to ensure all nodes can be connected to.
			type (
				DirExistStruct struct {
					dir   string
					exist bool
				}
			)
			var (
				msg string
			)

			hostDirsCache := make(map[string]DirExistStruct)
			dupErr := make(map[string]bool)
			for _, softwareGroup := range SoftwareGroupNames {
				if Terminated() {
					break
				}
				// Look up the software for each softwareGroup
				groupSoftware := upgradeconfig.GetGroupSoftware(softwareGroup)

				// Get the nodes for this group
				groupNodes := upgradeconfig.GetGroupNodes(softwareGroup)
				for _, node := range groupNodes {
					if Terminated() {
						break
					}
					if len(groupSoftware) == 0 {
						continue
					}
					for _, software := range groupSoftware {
						nodeInfo := upgradeconfig.GetNodeUpgradeInfo(node, software)
						sshConfig := softwareupgrade.NewSSHConfig(nodeInfo.SSHUserName, nodeInfo.SSHCert, node)
						for _, dirInfo := range nodeInfo.Copy {
							remoteDir := path.Dir(dirInfo.DestFilePath)
							hostDir := fmt.Sprintf("%s-%s", node, remoteDir)
							hostDirStruct := hostDirsCache[hostDir]
							if hostDirStruct == (DirExistStruct{}) {
								var err error
								hostDirStruct.dir = remoteDir
								switch action {
								case appActionAdd:
									{

										if hostDirStruct.exist, err = sshConfig.DirectoryExists(remoteDir); err == nil {
											if !hostDirStruct.exist {
												err = sshConfig.CreateDirectory(remoteDir)
												if err == nil {
													hostDirStruct.exist = true
												}
											}
											hostDirsCache[hostDir] = hostDirStruct
										}
									}
								case appActionUpgrade:
									{
										if hostDirStruct.exist, err = sshConfig.DirectoryExists(remoteDir); err == nil {
											hostDirsCache[hostDir] = hostDirStruct
											if !hostDirStruct.exist {
												msg = fmt.Sprintf("%sRemote directory: %s doesn't exist on node: %s\n",
													msg, remoteDir, node)
											}
										} else {
											errmsg := fmt.Sprintf("Node: %s error: %v", node, err)
											errExist := dupErr[errmsg]
											if !errExist {
												msg = fmt.Sprintf("%s%s\n", msg, errmsg)
												dupErr[errmsg] = true
											}
										}
									}
								}
							}
						}
					}
				}
			}
			if msg != "" {
				DebugLog.Println("Error(s) encountered in target directory verification.")
				DebugLog.Printf("%v", msg)
				return
			}
			if !Terminated() {
				DebugLog.Println("All remote directories verified.")
			}
		}
	}

	failedUpgradeInfo := softwareupgrade.NewFailedUpgradeInfo()
	rollbackSession := softwareupgrade.NewRollbackSession(rollbackSuffix)

	var resumeUpgrade bool
	defer func() {

		// If the failedUpgradeInfo structure isn't empty, there's a failure in upgrading
		// so save the information.
		if !failedUpgradeInfo.Empty() {
			data, err := json.Marshal(failedUpgradeInfo)
			if err == nil {
				softwareupgrade.SaveDataToFile(failedNodesFilename, data)
			} else {
				DebugLog.Println("Unable to marshal the failed upgrade information.")
			}
		}

		// Save the rollback data for either deletion, or rollback
		if !rollbackSession.RollbackInfo.Empty() {
			data, err := json.Marshal(rollbackSession)
			if err == nil {
				softwareupgrade.SaveDataToFile(rollbackInfoFilename, data)
			} else {
				DebugLog.Println("Unable to save the information for rollback.")
			}
		}

	}()

	switch action {
	case appActionAdd:
		{
			rollbackSession.Mode = mode
		}
	case appActionDeleteRollback:
		{
			if softwareupgrade.FileExists(rollbackInfoFilename) {
				data, err := softwareupgrade.ReadDataFromFile(rollbackInfoFilename)
				if err == nil {
					err = json.Unmarshal(data, &rollbackSession)
					if err != nil {
						// Clear the data so that it's not persisted again
						rollbackSession.RollbackInfo.Clear()
						failedUpgradeInfo.Clear()
						return
					}
					rollbackSuffix = rollbackSession.SessionSuffix
				}
			} else {
				DebugLog.Printf("Can't delete rollback as %s doesn't exist.\n", rollbackInfoFilename)
				return
			}
		}
	case appActionResumeUpgrade:
		{
			if !softwareupgrade.FileExists(failedNodesFilename) {
				DebugLog.Printf("Can't resume upgrade as %s doesn't exist.\n", failedNodesFilename)
				return
			}
			resumeUpgrade = true
		}
	case appActionRollback:
		{
			if softwareupgrade.FileExists(rollbackInfoFilename) {
				data, err := softwareupgrade.ReadDataFromFile(rollbackInfoFilename)
				if err == nil {
					err = json.Unmarshal(data, &rollbackSession)
					if err != nil {
						// Clear the data so that it's not persisted again
						rollbackSession.RollbackInfo.Clear()
						failedUpgradeInfo.Clear()
						return
					}
					rollbackSuffix = rollbackSession.SessionSuffix
				}
			} else {
				DebugLog.Printf("Can't rollback as %s doesn't exist\n", rollbackInfoFilename)
				return
			}
		}
	case appActionUpgrade:
		{
			// if a previous session exists...
			if softwareupgrade.FileExists(failedNodesFilename) {
				data, err := softwareupgrade.ReadDataFromFile(failedNodesFilename)
				if err == nil {
					err = json.Unmarshal(data, &failedUpgradeInfo.FailedNodeSoftware)
					resumeUpgrade = err == nil && len(failedUpgradeInfo.FailedNodeSoftware) > 0
				} else {
					DebugLog.Printf("Unable to read data from the failed nodes session due to error: %v", err)
				}
			} else {
				DebugLog.Println("Building node software list...")
				// Build the failedNodeSoftware list since this is a new session
				for _, softwareGroup := range SoftwareGroupNames {
					groupSoftware := upgradeconfig.GetGroupSoftware(softwareGroup)
					groupNodes := upgradeconfig.GetGroupNodes(softwareGroup)
					for _, node := range groupNodes {
						for _, software := range groupSoftware {
							failedUpgradeInfo.AddNodeSoftware(node, software)
						}
					}
				}
				DebugLog.Println("Node software list built.")
			}
		}
	}
	defer func() {
		// shows upgrade/rollback aborted/completed on app completion
		DebugLog.Println("%s %s", mode, appStatus)
	}()

	appStatus = "aborted"

	if Terminated() {
		return
	}

	for _, softwareGroup := range SoftwareGroupNames {
		// Look up the software for each softwareGroup
		groupSoftware := upgradeconfig.GetGroupSoftware(softwareGroup)

		// Get the nodes for this group
		groupNodes := upgradeconfig.GetGroupNodes(softwareGroup)
		if len(groupNodes) > 0 {
			var doPause bool
			DebugLog.Printf("Performing %s for software group: %s\n", mode, softwareGroup)
			for _, node := range groupNodes {
				if len(groupSoftware) == 0 {
					continue
				}
				if Terminated() {
					break
				}
				doPause = true
				for _, software := range groupSoftware {
					if Terminated() {
						break
					}

					// If this is a rollback, and the node and software doesn't exist
					// in the rollback data, then skip to the next one
					if action == appActionRollback {
						if !rollbackSession.RollbackInfo.ExistsNodeSoftware(node, software) {
							continue
						}
					}

					nodeInfo := upgradeconfig.GetNodeUpgradeInfo(node, software)

					// If this is a resume operation, and the node and software doesn't
					// exist in the failedUpgradeInfo then skip the current node and software.
					if resumeUpgrade {
						if !failedUpgradeInfo.ExistsNodeSoftware(node, software) {
							DebugLog.Println("Skipping software %s for node %s", software, node)
							continue
						}
					}

					// This message should be appropriate for different modes
					// It should be 1) Adding software %s to node %s
					//              2) Rolling back software %s for node %s
					//              3) Upgrading node %s with software %s
					//              4) Resuming upgrade for node %s with software %s
					//              5) Deleting software %s from node %s
					var actionMsg string
					switch action {
					case appActionAdd:
						{
							actionMsg = fmt.Sprintf("Adding software: %s to node: %s", software, node)
						}
					case appActionDeleteRollback:
						{
							actionMsg = fmt.Sprintf("Deleting rollback for software: %s from node: %s", software, node)
						}
					case appActionResumeUpgrade:
						{
							actionMsg = fmt.Sprintf("Resuming upgrade for node: %s with software: %s", node, software)
						}
					case appActionRollback:
						{
							actionMsg = fmt.Sprintf("Rolling back software: %s for node: %s", software, node)
						}
					case appActionUpgrade:
						{
							actionMsg = fmt.Sprintf("Upgrading node: %s with software: %s\n", node, software)
						}
					}
					DebugLog.Println(actionMsg)
					sshConfig := softwareupgrade.NewSSHConfig(nodeInfo.SSHUserName, nodeInfo.SSHCert, node)

					// Only stop the software if it's not Delete Rollback and not Add
					if action != appActionDeleteRollback && action != appActionAdd {
						// Stop the running software, upgrade it, then start the software
						StopCmd := nodeInfo.StopCmd
						StopResult, err := sshConfig.Run(StopCmd)
						if err != nil { // If stop failed, skip the upgrade!
							DebugLog.Printf(softwareupgrade.CNodeMsgSSS, node, softwareupgrade.CStop, err)
							continue
						}
						DebugLog.Printf(softwareupgrade.CNodeMsgSSS, node, softwareupgrade.CStop, StopResult)
					}

					if !dryRun {
						switch action {
						case appActionAdd:
							{
								err := nodeInfo.RunAdd(sshConfig)
								if err == nil {
									DebugLog.Println("Added software: %s to node: %s successfully", software, node)
								} else {
									DebugLog.Println("Failed to add software %s to node: %s", software, node)
								}
							}
						case appActionDeleteRollback:
							{
								err := nodeInfo.RunDeleteRollback(sshConfig, rollbackSuffix)
								if err != nil {
									DebugLog.Println("Failed to delete rollback for node: %s, software: %s due to %v", node, software, err)
								} else {
									DebugLog.Println("Deleted rollback for node: %s, software: %s", node, software)
								}
							}
						case appActionRollback:
							{

								err := nodeInfo.RunRollback(sshConfig, rollbackSuffix)
								if err != nil {
									DebugLog.Println("Rollback failed for node: %s, software: %s due to %v", node, software, err)
								} else {
									DebugLog.Println("Rolled back node: %s with software: %s successfully", node, software)
									rollbackSession.RollbackInfo.RemoveNodeSoftware(node, software)
								}
							}
						case appActionUpgrade:
							{
								err := nodeInfo.RunUpgrade(sshConfig) // the upgrade needs to either move or overwrite the older version
								if err != nil {
									DebugLog.Println("Error during RunUpgrade: %v", err)
								} else {
									DebugLog.Println("Upgraded node: %s with software %s successfully!", node, software)
									failedUpgradeInfo.RemoveNodeSoftware(node, software)
									rollbackSession.RollbackInfo.AddNodeSoftware(node, software)
								}
							}
						}
					}

					// Only start the software if it's not a delete rollback
					if action != appActionDeleteRollback && action != appActionAdd {
						StartCmd := nodeInfo.StartCmd
						StartResult, err := sshConfig.Run(StartCmd)
						if err != nil {
							DebugLog.Printf(softwareupgrade.CNodeMsgSSS, node, softwareupgrade.CStart, err)
							continue
						}
						DebugLog.Printf(softwareupgrade.CNodeMsgSSS, node, softwareupgrade.CStart, StartResult)
					}
				}
			}
			if Terminated() {
				break
			}
			if doPause { // pause only if upgrade has been run
				DebugLog.Printf("Pausing for %s...", upgradeconfig.Common.GroupPause)
				time.Sleep(upgradeconfig.Common.GroupPause.Duration)
				DebugLog.Println(" completed!")
				doPause = false // reset
			}
			DebugLog.Println("") // leave one line between one group and next group
		} else {
			DebugLog.Printf("No nodes for software group: %s\n", softwareGroup)
		}
	}

	if !Terminated() {
		appStatus = "completed"
	}
	softwareupgrade.ClearSSHConfigCache()
}

func main() {
	fmt.Println(softwareupgrade.CEximchainUpgradeTitle)

	rollbackSuffix = softwareupgrade.GetBackupSuffix()
	defaultRollbackName := fmt.Sprintf("~/Upgrade-Rollback-%s.session", rollbackSuffix)
	defaultFailedNodesFilename := fmt.Sprintf("~/Upgrade-Failed-%s.session", rollbackSuffix)

	flag.StringVar(&mode, "mode", "upgrade", "mode (add|resume-upgrade|upgrade|rollback|delete-rollback)")
	flag.BoolVar(&debug, "debug", false, "Specifies debug mode")
	flag.StringVar(&debugLogFilename, "debug-log", `~/Upgrade-debug.log`, "Specifies the debug log filename where logs are written to")
	flag.StringVar(&jsonFilename, "json", "", "Specifies the JSON configuration file to load nodes from")
	flag.StringVar(&failedNodesFilename, "failed-nodes", defaultFailedNodesFilename, "Specifes the file to load/save nodes that failed to upgrade")
	flag.StringVar(&rollbackInfoFilename, "rollback-filename", defaultRollbackName, "Specifies the rollback filename for this session")
	flag.BoolVar(&disableNodeVerification, "disable-node-verification", false, "Disables node IP resolution verification")
	flag.BoolVar(&disableFileVerification, "disable-file-verification", false, "Disables source file existence verification")
	flag.BoolVar(&disableTargetDirVerification, "disable-target-dir-verification", false, "Disables target directory existence verification")
	flag.BoolVar(&dryRun, "dry-run", true, "Enables testing mode, doesn't perform actual action, but starts and stops the software running on remote nodes")
	flag.Parse()

	switch strings.ToLower(mode) {
	case "add":
		{
			action = appActionAdd
		}
	case "delete", "delete-rollback":
		{
			action = appActionDeleteRollback
		}
	case "resume", "resume-upgrade":
		{
			action = appActionResumeUpgrade
		}
	case "rollback":
		{
			action = appActionRollback
		}
	case "upgrade":
		{
			action = appActionUpgrade
		}
	}

	// Ensures that JSONFilename is provided by user
	// and that mode must either be rollback or upgrade and that the given
	// JSON configuration file must exist
	if len(os.Args) <= 1 || jsonFilename == "" || !action.isValidAction() ||
		!softwareupgrade.FileExists(jsonFilename) {
		flag.PrintDefaults()
		return
	}

	if debug && debugLogFilename != "" {
		DebugLog.EnableDebug()
		if err := DebugLog.EnableDebugLog(debugLogFilename); err == nil {
			defer DebugLog.CloseDebugLog()
		} else {
			DebugLog.Println("Error: %v", err)
		}
	}

	DebugLog.Debugln(softwareupgrade.CEximchainUpgradeTitle)
	DebugLog.EnablePrintConsole()

	// Read JSON configuration file
	if expandedJSONFilename, err := softwareupgrade.Expand(jsonFilename); err == nil {
		jsonFilename = expandedJSONFilename
	} else {
		DebugLog.Println("Unable to interpret/parse %s due to %v", jsonFilename, err)
		return
	}

	if jsonContents, err := softwareupgrade.ReadDataFromFile(jsonFilename); err == nil {
		EnableSignalHandler()

		// Start processing the upgrade/rollback, etc...
		upgradeOrRollback(jsonContents)
		TerminateSignalHandler()
	} else {
		DebugLog.Println(`Error reading from JSON configuration file: "%s", error: %v`, jsonFilename, err)
	}

}
