package fsclient

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"crypto/md5"
	"encoding/hex"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	equal             = "="
	specialCharacters = "\t \n"
)

var (
	fsclientLog = ctrl.Log.WithName("FSClient")
	rootfspath  = os.Getenv("ROOT_FS_PATH")
	image       = os.Getenv("CRON_WRAPPER_IMAGE")
)

type CronEntryWithEnvvars struct {
	EnvVars   map[string]string
	Usercrons map[string]string
}

func Md5calc(stringtocalc string) string {
	hasher := md5.New()
	hasher.Write([]byte(stringtocalc))

	return hex.EncodeToString(hasher.Sum(nil))
}

func isAEnvVar(line string) bool {
	var cronregex = `^\s*\S+\s*=`
	re := regexp.MustCompile(cronregex)
	if re.MatchString(line) {
		return true
	} else {
		return false
	}
}

func GetFSEntryForUser(odpuser string) ([]CronEntryWithEnvvars, map[string]string) {
	var listOfAssociations []CronEntryWithEnvvars
	var usercrons = make(map[string]string)
	var allUsercrons = make(map[string]string)
	var isPreviousLineACommand = false
	var mapOfEnVars = make(map[string]string)
	var cronregex = `(^\s*#)|(^\s*$)`
	re := regexp.MustCompile(cronregex)

	pathtofile := rootfspath + odpuser + "/" + odpuser
	fsinfo, err := os.Stat(pathtofile)
	if errors.Is(err, os.ErrNotExist) {
		fsclientLog.Info("GetFSEntryForUser: File does not exist", " file:", pathtofile)
		return listOfAssociations, allUsercrons
	}
	if fsinfo.IsDir() {
		fsclientLog.Info("GetFSEntryForUser: This is a directory not a file", " Dir:", fsinfo.Name())
		return listOfAssociations, allUsercrons
	}

	userfile, err := os.Open(pathtofile)
	if err != nil {
		if os.IsPermission(err) {
			fsclientLog.Error(err, "GetFSEntryForUser: Cound not open file, No permission", " file:", userfile)
			return listOfAssociations, allUsercrons
		}
	}
	defer userfile.Close()

	scanner := bufio.NewScanner(userfile)
	for scanner.Scan() {
		line := scanner.Text()
		if !re.MatchString(line) {

			if isAEnvVar(line) {
				if isPreviousLineACommand {
					parts := strings.SplitN(line, equal, 2)
					key := strings.TrimSpace(parts[0])
					value := strings.TrimLeft(parts[1], specialCharacters)
					listOfAssociations = append(listOfAssociations, createAssociation(usercrons, mapOfEnVars))
					mapOfEnVars[key] = value
					usercrons = make(map[string]string)
					isPreviousLineACommand = false
				} else {
					parts := strings.SplitN(line, equal, 2)
					key := strings.TrimSpace(parts[0])
					value := strings.TrimLeft(parts[1], specialCharacters)
					mapOfEnVars[key] = value
					isPreviousLineACommand = false
				}
			} else {
			    var mapAsString = ConvertMapToString(mapOfEnVars)
				var unique = Md5calc(odpuser + line + image + mapAsString)
				usercrons[unique] = line
				allUsercrons[unique] = line
				isPreviousLineACommand = true
			}
		}
	}
	listOfAssociations = append(listOfAssociations, createAssociation(usercrons, mapOfEnVars))

	if err := scanner.Err(); err != nil {
		fsclientLog.Error(err, "GetFSEntryForUser: Could not read file ", " file:", userfile)
		return listOfAssociations, allUsercrons
	}

	return listOfAssociations, allUsercrons
}

func createAssociation(userCrons map[string]string, mapOfEnvVars map[string]string) CronEntryWithEnvvars {
	var mapOfEnvVarsCopy = make(map[string]string)
	for k, v := range mapOfEnvVars {
		mapOfEnvVarsCopy[k] = v
	}
	var usercronsCopy = make(map[string]string)
	for k, v := range userCrons {
		usercronsCopy[k] = v
	}
	return CronEntryWithEnvvars{mapOfEnvVarsCopy, usercronsCopy}
}

func ConvertMapToString(maptoConvert map[string]string) string {
	var stringBuilder strings.Builder
	//Iterations over maps in golang are randomized, sort the keys first for a consistent string
	keys := make([]string, 0, len(maptoConvert))
    for k := range maptoConvert {
       keys = append(keys, k)
    }
    sort.Strings(keys)
    for _, key := range keys {
        stringBuilder.WriteString(fmt.Sprintf("%s%s", key, maptoConvert[key]))
    }
	return stringBuilder.String()
}