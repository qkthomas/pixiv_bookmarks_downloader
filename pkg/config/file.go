// MIT License

// Copyright (c) [2022] [Lin Chen]

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	Config *configFile
)

const (
	configFileEnvName     = `PIXIV_DOWNLOADER_CONF`
	defaultConfigFilePath = `config.yaml`
)

type configFile struct {
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
	UserID   string `yaml:"UserID"`
	//some text for helping locate html nodes
	UsernameInputPH         string `yaml:"UsernameInputPlaceHolder"`
	PasswordInputPH         string `yaml:"PasswordInputPlaceHolder"`
	LoginButtonText         string `yaml:"LoginButtonText"`
	LogoutButtonText        string `yaml:"LogoutButtonText"`
	ConfirmLogoutButtonText string `yaml:"ConfirmLogoutButtonText"`
	BookmarkAnchorText      string `yaml:"BookmarkAnchorText"`
}

func init() {
	path := os.Getenv(configFileEnvName)
	if path == "" {
		path = defaultConfigFilePath
	}
	c, err := readConfigFile(path)
	if err != nil {
		panic(err.Error())
	}
	Config = &c
}

func readConfigFile(path string) (c configFile, err error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return c, fmt.Errorf("unable to read file at \"%s\": %+v", path, err)
	}
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return c, fmt.Errorf("unable to unmarshal yaml file at \"%s\": %+v", path, err)
	}

	return c, nil
}
