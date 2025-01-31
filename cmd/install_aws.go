/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"gov.gsa.fac.cgov-util/internal/environments"
	"gov.gsa.fac.cgov-util/internal/logging"
	"gov.gsa.fac.cgov-util/internal/util"
)

func CheckInstallation() {
	checkinstallation := []string{
		util.AWS_path,
		"--version",
	}
	command := strings.Join(checkinstallation[:], " ")
	check := exec.Command("bash", "-c", command)
	checkOutput, checkError := check.Output()
	util.ErrorCheck(string(checkOutput), checkError)
}

func InstallAWS() {
	logging.Logger.Printf("ENV detected to be a cloud.gov environment. Installing AWS CLI.")
	// curl -x $https_proxy -L "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
	getaws := []string{
		"curl",
		"-x",
		os.Getenv("https_proxy"),
		"-L",
		"https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip",
		"-o",
		"awscliv2.zip",
	}
	curlCommand := strings.Join(getaws[:], " ")
	logging.Logger.Printf("Fetching aws cli via curl...")
	logging.Logger.Printf("Running command: " + curlCommand)
	curl := exec.Command("bash", "-c", curlCommand)
	curlOutput, curlError := curl.Output()
	util.ErrorCheck(string(curlOutput), curlError)

	// unzip awscliv2.zip && rm awscliv2.zip
	unzipaws := []string{
		"unzip",
		"-q",
		"awscliv2.zip",
		"&&",
		"rm",
		"awscliv2.zip",
	}
	unzipCommand := strings.Join(unzipaws[:], " ")
	logging.Logger.Printf("Unzipping aws cli...")
	logging.Logger.Printf("Running command: " + unzipCommand)
	extract := exec.Command("bash", "-c", unzipCommand)
	unzipOutput, unzipError := extract.Output()
	util.ErrorCheck(string(unzipOutput), unzipError)

	// ./aws/install -i ~/usr -b ~/bin
	installaws := []string{
		"./aws/install",
		"-i",
		"~/usr",
		"-b",
		"~/bin",
	}
	installCommand := strings.Join(installaws[:], " ")
	logging.Logger.Printf("Installing aws to bin...")
	logging.Logger.Printf("Running command: " + installCommand)
	install := exec.Command("bash", "-c", installCommand)
	installOutput, installError := install.Output()
	util.ErrorCheck(string(installOutput), installError)

	// Regardless of the case, check to see if AWS-CLI is installed or not.
	CheckInstallation()
}

// installAwsCmd represents the installAws command
var installAwsCmd = &cobra.Command{
	Use:   "install_aws",
	Short: "Install AWS-CLI",
	Long:  `This command will curl the necessary aws-cli package and install it`,
	Run: func(cmd *cobra.Command, args []string) {
		if slices.Contains([]string{environments.SANDBOX, environments.DEVELOPMENT, environments.PREVIEW, environments.STAGING, environments.PRODUCTION}, os.Getenv("ENV")) {
			InstallAWS()
		} else {
			logging.Logger.Printf("ENV set to LOCAL or TESTING, aws-cli is not necessary to install.")
			// Regardless of the case, check to see if AWS-CLI is installed or not.
			CheckInstallation()
		}
	},
}

func init() {
	rootCmd.AddCommand(installAwsCmd)
}
