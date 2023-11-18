How to build
==

1. In order to build this, download the binary release for [Go](https://golang.org/dl/) for your operating system, install it, and make it available on your environment's PATH.
2. Open a terminal and go to the directory where the repository is located.
3. Ensure build.sh is executable. Run:
    1.  chmod a+x build.sh
4. Update the GOPATH environment variable in vars.sh
5. Run the build.sh script, which generates the CreateConfig and Upgrade executables.
6. Run CreateConfig like so:
    1. CreateConfig template-filename terraform-output-json-filename outputfilename, eg, CreateConfig ~/template.json ~/terraformoutput.json ~/Upgrade.json
7. Run Upgrade with any necessary parameters, like so:
    1. Upgrade -debug-log ~/Upgrade-debug.log -json ~/Upgrade.json
