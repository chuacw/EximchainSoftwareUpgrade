Table of Contents
=================
   * [Introduction](#introduction)
   * [CreateConfig](#createconfig)
        * [Command line parameters](#createconfig-command-line-parameters)
        * [Template format](#createconfig-template-format)
   * [CreateGraph](#creategraph)
        * [Command line parameters](#creategraph-command-line-parameters)
        * [Example](#creategraph-example)
   * [Upgrade](#upgrade)
        * [command line parameters](#upgrade-command-line-parameters)
        * [JSON configuration file format](#json-configuration-file-format)
        * [Troubleshooting](#troubleshooting)
  

Introduction
==
This is a suite of two software (CreateConfig and Upgrade) written to help in upgrading, as well as adding software onto Eximchain nodes.

In order to run this, you'll need to build it.

To build this, you'll need to refer instructions given in the BUILD.md.

CreateConfig
==

CreateConfig is a tool created to transform Terraform's json output into a configuration suitable to be used for the upgrade tool. To get Terraform's output in JSON format, run the following command:

```
terraform output -json > terraformoutput.txt
```
The above command causes Terraform to output its data is JSON format into the file terraformoutput.txt. You can use any filename.

Examples of the CreateConfig invocation is as below:
```

./CreateConfig -mode=aws -output1="Key=NetworkId,Value=84826,Key=Role,Value=Bootnode;bootnode_ips;map;PublicIpAddress" -output2="Key=NetworkId,Value=84826,Key=Role,Value=Maker;quorum_maker_node_dns;map;PublicDnsName" -output3="Key=NetworkId,Value=84826,Key=Role,Value=Observer;quorum_observer_node_dns;map;PublicDnsName" -output4="Key=NetworkId,Value=84826,Key=Role,Value=Validator;quorum_validator_node_dns;map;PublicDnsName" -output5="Key=NetworkId,Value=84826,Key=Role,Value=Vault;vault_server_ips;list;PublicIpAddress" -output=/tmp/terraform-AWS.json -sak=FOaR4xxxxxxxx -akid=AKIAxxxxxxxxxx -region=us-east-1

./CreateConfig -mode=aws -output1="Key=NetworkId,Value=84826,Key=Role,Value=Bootnode;bootnode_ips;map;PublicIpAddress" -output2="Key=NetworkId,Value=84826,Key=Role,Value=Maker;quorum_maker_node_dns;map;PublicDnsName" -output3="Key=NetworkId,Value=84826,Key=Role,Value=Observer;quorum_observer_node_dns;map;PublicDnsName" -output4="Key=NetworkId,Value=84826,Key=Role,Value=Validator;quorum_validator_node_dns;map;PublicDnsName" -output5="Key=NetworkId,Value=84826,Key=Role,Value=Vault;vault_server_ips;list;PublicIpAddress" -output=/tmp/terraform-AWS.json 

CreateConfig -mode=cli -input=template.json -output=terraformoutput.txt -terraform-json=upgrade.json -remove-quote=true -remove-delimiter=true

CreateConfig -mode=tfe -input=template.json -output=terraformoutput.txt -workspace=workspace-name -organization=eximchain -auth=authtoken
```

In aws mode, CreateConfig reads data from AWS based on credentials and region read from ~/.aws/credentials and ~/.aws/config or from the command line. The difference between the first and second example is that the credentials and region are specified on the command line.

In cli mode, CreateConfig reads from a file (which is the redirected output of terraform output -json) and combines it with the specified template file to produce the output file.

In tfe mode, CreateConfig uses the authorization token to connect to Terraform Enterprise and retrieve the given organization's workspace's latest run's output and combines it with the specified template file to produce the output file.

CreateConfig command line parameters
==
* 	-template - Filename of template
*	-terraform-json - Filename of Terraform output in JSON format (not required in TFE mode)
*   -output - Filename to write output to
*   -remove-quote - true|false, remove quotes from output
*   -remove-delimiter - remove commas from output separating items
*   -mode - aws|cli|tfe, command line interface, or Terraform Enterprise API integration
    *   in AWS mode, these parameters are required: outputN where N is 1 to 10, and output
    *   in cli mode, these parameters are required: output, template, terraform-json
    *   in tfe mode, these parameters are required: auth, organization, workspace
*   -workspace - Name of workspace (only for tfe mode)
*   -organization - Name of organization (only for tfe mode)
*   -auth - Authorization token (only for tfe mode)
*   -akid - AWS Access Key ID (only for aws mode)
*   -sak - AWS Secret Access Key (only for aws mode)
*   -region - default region (only for aws mode)
*   -debug - in AWS mode, useful for looking at the filter, and progress.


CreateConfig template format
==
The template format is a left brace, and a percent, "{%", followed by the node name, followed by a right brace, "}".

An example template file looks like this:
```
{%bootnode_ips} {%vault_server_ips}
```

And if the Terraform output in JSON format contains this:
```
{
    "bootnode_ips": {
        "sensitive": false,
        "type": "map",
        "value": {
            "ap-northeast-1": [],
            "ap-northeast-2": [],
            "ap-south-1": [],
            "ap-southeast-1": [],
            "ap-southeast-2": [],
            "ca-central-1": [],
            "eu-central-1": [],
            "eu-west-1": [],
            "eu-west-2": [],
            "sa-east-1": [],
            "us-east-1": [
                "18.207.120.214"
            ],
            "us-east-2": [
                "18.222.37.52",
                "18.188.115.226"
            ],
            "us-west-1": [],
            "us-west-2": []
        }
    },
    "vault_server_ips": {
        "sensitive": false,
        "type": "list",
        "value": [
            "54.175.210.140"
        ]
    }
}
```

and given that -remove-delimiter=true, then the output would be the combined values of the nodes:
```
18.207.120.214 18.222.37.52 18.188.115.226 54.175.210.140
```

CreateGraph
==
CreateGraph is a tool used to visualize Eximchain nodes by generating a DOT output file.
It reads either a single input file, or a file containing a list of files, each of which contains the remote addresses that a node is connected to.
It then generates the output file using the [DOT](https://emden.github.io/_pages/doc/info/lang.html) language. The generated file can then be opened by an application that implements the GraphViz implementation, or [viewed online](http://www.webgraphviz.com/) (by copying and pasting the contents of the output file).

CreateGraph command line parameters
==
*   -concurrent - number of iterations to run concurrently (minimum and default of 1)
*   -extension - The file extension of the input files
* 	-in - Name of input file containing remote addresses.
*   -iterations - number of iterations to run (default of 100)
*	-list - Name of file containing list of input files containing remote addresses. (use either -in or -list but not both)
*   -output - Filename to write output to
*   -radius - Whether to calculate the radius or not (true|false)

CreateGraph example
==
Here's an example on how to call it:
```
./CreateGraph -concurrent=100 -extension=.json.raw -list=~/nodelist.txt -out=~/EximchainNodes.dot
```

This example assumes that nodelist.txt is a list of files.

Example nodelist.txt:
```
ec2-13-232-223-9.ap-south-1.compute.amazonaws.com,13.232.223.9,10.0.56.87
ec2-13-127-26-148.ap-south-1.compute.amazonaws.com,13.127.26.148,10.0.57.98
ec2-13-232-248-251.ap-south-1.compute.amazonaws.com,13.232.248.251,10.0.56.85
ec2-52-66-33-35.ap-south-1.compute.amazonaws.com,52.66.33.35,10.0.57.17
ec2-13-233-1-173.ap-south-1.compute.amazonaws.com,13.233.1.173,10.0.56.188
ec2-18-209-9-107.compute-1.amazonaws.com,18.209.9.107,10.0.0.84
```

Example ec2-13-232-223-9.ap-south-1.compute.amazonaws.com.json.raw
```
Welcome to the Geth JavaScript console!

instance: Geth/v1.5.0-unstable-6729e9a5/linux/go1.10.3
coinbase: 0x1f9b5c63f7395e32261bd8601a69c70a8c5e8303
at block: 150 (Wed, 05 Sep 2018 07:20:06 UTC)
 datadir: /home/ubuntu/.ethereum
 modules: admin:1.0 debug:1.0 eth:1.0 net:1.0 personal:1.0 quorum:1.0 rpc:1.0 txpool:1.0 web3:1.0

> 
[{
    caps: ["eth/62", "eth/63"],
    id: "01d5a25c829939e931b9e56bfc66972fbfed5f18f593cac00745986c9dddcf878c2c9fac81bc0b55c4c0fc8af14cb0d5910821697b6546f1b7d81d2642559141",
    name: "Geth/v1.5.0-unstable-6729e9a5/linux/go1.10.3",
    network: {
      localAddress: "10.0.9.238:33980",
      remoteAddress: "54.197.13.5:21000"
    },
    protocols: {
      eth: {
        difficulty: 20247432,
        head: "0x8ef211d535cc692b085eceed68a834f4f45316e4f0c957adac80ffb15a30414d",
        version: 63
      }
    }
}, {
    caps: ["eth/62", "eth/63"],
    id: "11f206e5d15a17959f60b595a9c929161125d37ff988b88ca9bbf025da7bfecb48d478f68dfc642ba5a397a465df27d92d3b25cad70a0710a954ffe3e7aaf12f",
    name: "Geth/v1.5.0-unstable-6729e9a5/linux/go1.10.3",
    network: {
      localAddress: "10.0.9.238:21000",
      remoteAddress: "54.197.84.79:47130"
    },
    protocols: {
      eth: {
        difficulty: 20106613,
        head: "0xb6a3f8aaa501ce95e088089fb10b83a3fbc0c8c724dd2ee1f7cec69ab402d3f4",
        version: 63
      }
    }
}, {
    caps: ["eth/62", "eth/63"],
    id: "1e5973dfc067258e01777320ab1f3d7db894a79faefea9e7d116447a5d51ce4c83d0c8c68abdf1b101b6e796b99f573b6114c9cc5c0fa15998eb6706d655fb9e",
    name: "Geth/v1.5.0-unstable-6729e9a5/linux/go1.10.3",
    network: {
      localAddress: "10.0.9.238:21000",
      remoteAddress: "35.172.141.185:55522"
    },
    protocols: {
      eth: {
        difficulty: 19684564,
        head: "0x23fbc1d872f2a154b64cc56ff4b5537c10f15b1d4709402db8453b475c51e1e1",
        version: 63
      }
    }
}, {
    caps: ["eth/62", "eth/63"],
    id: "21116d4efc4a342591ca7dad6eb6f36ba14b81841d2f662d9ce0ba300b13bd230b05e75ab0531814d53b4ffb980b59338f2937fb153a614330ffa8eeb6ec94fd",
    name: "Geth/v1.5.0-unstable-6729e9a5/linux/go1.10.3",
    network: {
      localAddress: "10.0.9.238:60650",
      remoteAddress: "18.209.9.107:21000"
    },
    protocols: {
      eth: {
        difficulty: 20247432,
        head: "0x8ef211d535cc692b085eceed68a834f4f45316e4f0c957adac80ffb15a30414d",
        version: 63
      }
    }
}, {
    caps: ["eth/62", "eth/63"],
    id: "22fb8749df8f3428e8085062af3af20065ca4696fff6f242415aa6520adba9a7d4f875123abc3d0c595ac6684fb501890814f0d3321979ee52ae56ba92f1bade",
    name: "Geth/v1.5.0-unstable-6729e9a5/linux/go1.10.3",
    network: {
      localAddress: "10.0.9.238:21000",
      remoteAddress: "54.210.124.111:40806"
    },
    protocols: {
      eth: {
        difficulty: 20247432,
        head: "0x8ef211d535cc692b085eceed68a834f4f45316e4f0c957adac80ffb15a30414d",
        version: 63
      }
    }
}, {
    caps: ["eth/62", "eth/63"],
    id: "2d9cb831f0a62e865220313ed62c800db871a40b2c76d734bf375bc02b149c597ffcb170245dde55a857941776aab97628efa931e87d09aa83bbd00123f62b2a",
    name: "Geth/v1.5.0-unstable-6729e9a5/linux/go1.10.3",
    network: {
      localAddress: "10.0.9.238:50040",
      remoteAddress: "52.15.135.96:21000"
    },
    protocols: {
      eth: {
        difficulty: 20247432,
        head: "0x8ef211d535cc692b085eceed68a834f4f45316e4f0c957adac80ffb15a30414d",
        version: 63
      }
    }
}]
> 
```

Example output file:
```
graph EximchainNodes {
  subgraph cluster0 {
    node [color=white];
    label="Eximchain Nodes, min cut=14, IPs=45, diameter=6"
    "     ";
  };
  subgraph cluster_1 {
    label="Subgraph 1"
    "13.232.223.9";
    "54.197.13.5";
    "54.197.84.79";
    "52.66.139.183";
    "13.232.142.88";
    "13.127.26.148";
  };
}
```

Upgrade
==

Upgrade is a tool created to upgrade and add software to the target Eximchain nodes.

Upgrade command line parameters
==

* -debug Specifies debug mode - true|false, when this is specified, more debug information go into the debug log.
* -debug-log logfilename - specifies the name of the debug log to write to.
* -disable-file-verification - true|false, disables source file existence verification.
* -disable-target-dir-verification - true|false, disables target directory existence verification.
* -dry-run - true|false, enables testing mode, doesn't perform actual action, but starts and stops the software running on remote nodes
* -failed-nodes - Specifies the filename to load/save nodes that failed to upgrade.
* -json jsonfilename - specifies the name of the JSON configuration file to read from. This must always be present.
* -mode - Specifies the operating mode - add, delete-rollback, resume-upgrade, rollback, upgrade (default: upgrade)
* -rollback-filename - Specifies the rollback filename for this session.
  * Mode: add, adds the specified software in the configuration to the target nodes.
  * Mode: delete-rollback, removes the rollback files on the target nodes (only for software upgraded, not for software added)
  * Mode: resume-upgrade, continues the previous upgrade.
  * Mode: rollback, the files specified in this session will be used to remove the upgraded software on the target nodes.
  * Mode: upgrade, upgrade the software on the target nodes.
* -help - brings up information about the parameters.

Example
```
    -json=~/Documents/GitHub/SoftwareUpgrade/LaunchUpgrade.json -debug=true -debug-log=~/EximchainUpgrade.log
```

This launches the upgrader telling it to read the upgrade information from the file LaunchUpgrade.json, and to enable debug log output to the EximchainUpgrade.log file in the user home directory.

The rollback-filename parameter allows target nodes to rollback to the state they were before being upgraded.

JSON configuration file format
==

The JSON configuration file is a JSON object, which consists of a number of objects (Three objects at the moment). 
1. The software object defines multiple objects that specifies start, stop and files to copy to the target node, as well as commands to execute on the target node. 

2. The common object defines the location of the ssh certificate, the username to use.

3. The groupnodes object lists nodes belonging to groups listed in the software object's child nodes.

The software object defines the software objects that are to be upgraded/added on the target nodes.
Each child software object can be arbitrarily named (**___the same names must be used in the groupnodes children nodes___**). It has a start and a stop string, and a copy object. 

The start and stop string specifies commands to execute, in order to start and stop the software being upgraded.
The stop command is executed first.
Each Copy object has numbered objects starting from 0, or 1. Each numbered object has a Local_Filename, Remote_Filename, and a Permissions string.
The Local_Filename string specifies the filename of the file to copy from. The Remote_Filename specifies the destination on the target node to copy the file to. The Permissions string specifies the ownership of the copied file, and is applied after the file has been copied over to the target node.
After all numbered objects are copied, the start command is then executed.

Table of child software object properties.

| Property | Type | Description |
|---|---|---|
| start  	| string  	| The command to execute, in order to start the software after being added/upgraded.  	|
| stop  	| string  	| The command to execute, in order to stop the software before being upgraded. May be empty if the software is to be added. 	|
| Copy  	| object  	| The file(s) to copy, in order to add/upgrade the software to/on the target node.  	|

Table of Copy object properties.

| Property | Type | Description |
|---|---|---|
| Local_Filename  	| string  	| Full path to the file to copy.  	|
| Remote_Filename  	| string  	| Full path on the target node for the file to be copied to.  	|
| Permissions  	| string  	| A 4-digit permissions string.  	|
| preupgrade  	| array of strings  	| Command(s) to execute before the upgrade starts. If empty, no commands are executed. 	|
| postupgrade  	| array of strings  	| Command(s) to execute after the upgrade is completed. If empty, no commands are executed. 	|

Table of common object properties.

| Property | Type | Description |
|---|---|---|
| ssh_cert  	| string  	| Filename of the SSH certificate used to SSH to target nodes.  	|
| ssh_username  	| string  	| Username used to SSH to target nodes.  	|
| group_pause_after_upgrade  	| string  	| Specifies the amount of time to delay after upgrading a software group. 1h5m3s would mean 1 hour 5 minute and 3 seconds. The amount of time to delay is specified using this nomenclature. 	|
| software_group  	| array of strings  	| Specifies the list of software that comprised this group. The software names used must be the same as those listed under the top level software object.  	|

Table of groupnode properties.

| Property | Type | Description |
|---|---|---|
| Same name as used under the top-level software object. | array of strings | Specifies the hostname of the target nodes.| 

An example of the JSON configuration file format follows.

```
{
    "software": {
        "blockmetrics": {
            "start": "sudo supervisorctl start blockmetrics",
            "stop": "sudo supervisorctl stop blockmetrics",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/block-metrics.py",
                    "Remote_Filename": "/opt/quorum/bin/block-metrics.py",
                    "Permissions": "0644"
                }
            }
        },
        "bootnode": {
            "start": "sudo supervisorctl start bootnode",
            "stop": "sudo supervisorctl stop bootnode",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/bootnode",
                    "Remote_Filename": "/usr/local/bin/bootnode",
                    "Permissions": "0755"
                }
            }
        },
        "cloudwatchmetrics": {
            "start": "sudo supervisorctl start cloudwatchmetrics",
            "stop": "sudo supervisorctl stop cloudwatchmetrics",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/cloudwatch-metrics.sh",
                    "Remote_Filename": "/opt/quorum/bin/cloudwatch-metrics.sh",
                    "Permissions": "0644"
                }
            }
        },
        "constellation": {
            "start": "sudo supervisorctl start constellation",
            "stop": "sudo supervisorctl stop constellation",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/constellation-node",
                    "Remote_Filename": "/usr/local/bin/constellation-node",
                    "Permissions": "0755"
                }
            }
        },
        "consul": {
            "start": "sudo supervisorctl start consul",
            "stop": "sudo supervisorctl stop consul",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/consul",
                    "Remote_Filename": "/opt/consul/bin/consul",
                    "Permissions": "0755"
                }
            }
        },
        "crashconstellation": {
            "start": "sudo supervisorctl start crashconstellation",
            "stop": "sudo supervisorctl stop crashconstellation",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/crashcloudwatch.py",
                    "Remote_Filename": "/opt/quorum/bin/crashcloudwatch.py",
                    "Permissions": "0744"
                }
            }
        },
        "crashquorum": {
            "start": "sudo supervisorctl start crashquorum",
            "stop": "sudo supervisorctl stop crashquorum"
        },
        "quorum": {
            "start": "sudo supervisorctl start quorum",
            "stop": "sudo supervisorctl stop quorum",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/geth",
                    "Remote_Filename": "/usr/local/bin/geth",
                    "Permissions": "0755"
                }
            }
        },
        "vault": {
            "preupgrade": [""],
            "postupgrade": [""],
            "start": "sudo supervisorctl start vault",
            "stop": "sudo supervisorctl stop vault",
            "Copy": {
                "1": {
                    "Local_Filename": "/tmp/vault",
                    "Remote_Filename": "/opt/vault/bin/vault",
                    "Permissions": "0755"
                }
            }
        }
    },
    "common": {
        "ssh_cert": "~/.ssh/quorum",
        "ssh_timeout": "1m",
        "ssh_username": "ubuntu",
        "group_pause_after_upgrade": "6m15s",
        "software_group": {
            "Quorum-Makers": [
                "blockmetrics",
                "consul",
                "constellation",
                "quorum"
            ],
            "Quorum-Observers": [
                "blockmetrics",
                "cloudwatchmetrics",
                "constellation",
                "consul",
                "crashconstellation",
                "crashquorum",
                "quorum"
            ],
            "Quorum-Validators": [
                "blockmetrics",
                "cloudwatchmetrics",
                "constellation",
                "consul",
                "crashconstellation",
                "crashquorum",
                "quorum"
            ],
            "Bootnodes": [
                "constellation",
                "consul",
                "bootnode"
            ],
            "VaultServers": [
                "consul",
                "vault"
            ]
        }
    },
    "groupnodes": {
        "Quorum-Makers": [
            "ec2-54-164-95-40.compute-1.amazonaws.com",
            "name3",
            "name4"
        ],
        "Quorum-Observers": [
            "ec2-52-201-244-132.compute-1.amazonaws.com",
            "name5",
            "name6"
        ],
        "Quorum-Validators": [
            "ec2-52-72-195-7.compute-1.amazonaws.com",
            "name7",
            "name8",
            "name9"
        ],
        "Bootnodes": [
            "ec2-54-166-128-218.compute-1.amazonaws.com",
            "name10..."
        ],
        "VaultServers": [
            "34.228.16.117",
            "moreIP, or DNS"
        ]
    }
}
```

The "Quorum-Makers" in "software_group" specifies that it consists of the "blockmetrics", "consul", "constellation" and "quorum" software.
The "Quorum-Makers" in "groupnodes" specifies that the hostnames are: "ec2-54-164-95-40.compute-1.amazonaws.com", "name3", "name4", and that the software in the "Quorum-Makers" in "software_group" will be deployed to these hostnames. _The software group name used in "software_group" and "groupnodes" must be the same, so that the application knows that the software specified in the "software_group" is to be deployed to the nodes specified in the "groupnodes" under the same name._

Troubleshooting
==
By default, this software produces a debug log called Upgrade-debug.log at ~/, unless it is disabled.

Any errors should appear in the debug log.
