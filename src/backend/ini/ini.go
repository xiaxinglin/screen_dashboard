package iniparser

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type _Section map[string]string

type IniFile struct {
	fpath    string
	sections map[string]_Section
}

func (self *IniFile) GetString(section string, name string) (string, bool) {
	s := self.sections[section]
	if s == nil {
		return "", false
	}
	v, ok := s[name]
	return v, ok
}

func (self *IniFile) GetInt(section string, name string) (int, bool) {
	v, err := strconv.Atoi(self.sections[section][name])
	if err != nil {
		return 0, false
	}
	return v, true
}

func (self *IniFile) GetFloat(section string, name string) (float64, bool) {
	v, err := strconv.ParseFloat(self.sections[section][name], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (self *IniFile) parse(fpath string, encoding string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()
	self.sections = make(map[string]_Section)
	scanner := bufio.NewScanner(f)
	reSection := regexp.MustCompile(`^\[(.+?)\]$`)
	reAssign := regexp.MustCompile(`^([^=]+?)=(.+?)$`)
	var section _Section = nil
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' || line[0] == ';' {
			continue
		}
		if g := reSection.FindStringSubmatch(line); g != nil {
			name := strings.TrimSpace(g[1])
			if name != "" {
				section = make(_Section)
				self.sections[name] = section
			}
		} else if g := reAssign.FindStringSubmatch(line); g != nil {
			key := strings.TrimSpace(g[1])
			value := strings.TrimSpace(g[2])
			section[key] = value
		}
	}
	return nil
}

func LoadFile(fpath string, encoding string) (*IniFile, error) {
	ini := &IniFile{}
	err := ini.parse(fpath, encoding)
	if err != nil {
		return nil, err
	}
	return ini, nil
}