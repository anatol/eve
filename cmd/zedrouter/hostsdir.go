// Copyright (c) 2017 Zededa, Inc.
// All rights reserved.

// hostsdir configlet for overlay interface towards domU

package zedrouter

import (
	"fmt"
	"github.com/zededa/go-provision/types"
	"log"
	"net"
	"os"
)

// Create the hosts file for the overlay DNS resolution
// Would be more polite to return an error then to Fatal
func createHostsConfiglet(cfgDirname string, nameToEidList []types.NameToEid) {
	if debug {
		log.Printf("createHostsConfiglet: dir %s nameToEidList %v\n",
			cfgDirname, nameToEidList)
	}
	ensureDir(cfgDirname)

	for _, ne := range nameToEidList {
		addIPToHostsConfiglet(cfgDirname, ne.HostName, ne.EIDs)
	}
}

func ensureDir(dirname string) {
	if _, err := os.Stat(dirname); err != nil {
		err := os.Mkdir(dirname, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Create one file per hostname
func addIPToHostsConfiglet(cfgDirname string, hostname string, addrs []net.IP) {
	ensureDir(cfgDirname)
	cfgPathname := cfgDirname + "/" + hostname
	file, err := os.Create(cfgPathname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	for _, addr := range addrs {
		file.WriteString(fmt.Sprintf("%s	%s\n",
			addr.String(), hostname))
	}
}

// Create one file per hostname
func addToHostsConfiglet(cfgDirname string, hostname string, addrs []string) {
	ensureDir(cfgDirname)
	cfgPathname := cfgDirname + "/" + hostname
	file, err := os.Create(cfgPathname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	for _, addr := range addrs {
		file.WriteString(fmt.Sprintf("%s	%s\n", addr, hostname))
	}
}

func removeFromHostsConfiglet(cfgDirname string, hostname string) {
	cfgPathname := cfgDirname + "/" + hostname
	if err := os.Remove(cfgPathname); err != nil {
		log.Println("removeFromHostsConfiglet: ", err)
	}
}

func containsHostName(nameToEidList []types.NameToEid, hostname string) bool {
	for _, ne := range nameToEidList {
		if hostname == ne.HostName {
			return true
		}
	}
	return false
}

func containsEID(nameToEidList []types.NameToEid, EID net.IP) bool {
	for _, ne := range nameToEidList {
		for _, eid := range ne.EIDs {
			if eid.Equal(EID) {
				return true
			}
		}
	}
	return false
}

func updateHostsConfiglet(cfgDirname string,
	oldNameToEidList []types.NameToEid, newNameToEidList []types.NameToEid) {
	if debug {
		log.Printf("updateHostsConfiglet: dir %s old %v, new %v\n",
			cfgDirname, oldNameToEidList, newNameToEidList)
	}
	// Look for hosts which should be deleted
	for _, ne := range oldNameToEidList {
		if !containsHostName(newNameToEidList, ne.HostName) {
			cfgPathname := cfgDirname + "/" + ne.HostName
			if err := os.Remove(cfgPathname); err != nil {
				log.Println("updateHostsConfiglet: ", err)
			}
		}
	}

	for _, ne := range newNameToEidList {
		cfgPathname := cfgDirname + "/" + ne.HostName
		file, err := os.Create(cfgPathname)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		for _, eid := range ne.EIDs {
			file.WriteString(fmt.Sprintf("%s	%s\n",
				eid, ne.HostName))
		}
	}
}

func deleteHostsConfiglet(cfgDirname string, printOnError bool) {
	if debug {
		log.Printf("deleteHostsConfiglet: dir %s\n", cfgDirname)
	}
	err := os.RemoveAll(cfgDirname)
	if err != nil && printOnError {
		log.Println("deleteHostsConfiglet: ", err)
	}
}
