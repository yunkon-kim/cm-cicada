package config

import (
	"errors"
	"fmt"
	"github.com/cloud-barista/cm-cicada/common"
	"github.com/jollaman999/utils/fileutil"
	"gopkg.in/yaml.v3"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type cmCicadaConfig struct {
	CMCicada struct {
		AirflowServer struct {
			Address       string `yaml:"address"`
			UseTLS        string `yaml:"use_tls"`
			SkipTLSVerify string `yaml:"skip_tls_verify"`
			Timeout       string `yaml:"timeout"`
			Username      string `yaml:"username"`
			Password      string `yaml:"password"`
		} `yaml:"airflow-server"`
		Listen struct {
			Port string `yaml:"port"`
		} `yaml:"listen"`
	} `yaml:"cm-cicada"`
}

var CMCicadaConfig cmCicadaConfig
var cmCicadaConfigFile = "cm-cicada.yaml"

func checkCMCicadaConfigFile() error {
	if CMCicadaConfig.CMCicada.AirflowServer.Address == "" {
		return errors.New("config error: cm-cicada.airflow-server.address is empty")
	}

	addrSplit := strings.Split(CMCicadaConfig.CMCicada.AirflowServer.Address, ":")
	if len(addrSplit) < 2 {
		return errors.New("config error: invalid cm-cicada.airflow-server.address must be {IP or IPv6 or Domain}:{Port} form")
	}
	port, err := strconv.Atoi(addrSplit[len(addrSplit)-1])
	if err != nil || port < 1 || port > 65535 {
		return errors.New("config error: cm-cicada.airflow-server.address has invalid port value")
	}
	addr, _ := strings.CutSuffix(CMCicadaConfig.CMCicada.AirflowServer.Address, ":"+strconv.Itoa(port))
	_, err = netip.ParseAddr(addr)
	if err != nil {
		_, err = net.LookupIP(addr)
		if err != nil {
			return errors.New("config error: cm-cicada.airflow-server.address has invalid address value " +
				"or can't find the domain (" + addr + ")")
		}
	}

	useTLS, err := strconv.ParseBool(strings.ToLower(CMCicadaConfig.CMCicada.AirflowServer.UseTLS))
	if err != nil {
		return errors.New("config error: cm-cicada.airflow-server.use_tls has invalid value")
	}
	if useTLS {
		_, err = strconv.ParseBool(strings.ToLower(CMCicadaConfig.CMCicada.AirflowServer.SkipTLSVerify))
		if err != nil {
			return errors.New("config error: cm-cicada.airflow-server.skip_tls_verify has invalid value")
		}
	}

	if CMCicadaConfig.CMCicada.AirflowServer.Timeout == "" {
		return errors.New("config error: cm-cicada.airflow-server.timeout is empty")
	}
	timeout, err := strconv.Atoi(CMCicadaConfig.CMCicada.AirflowServer.Timeout)
	if err != nil || timeout < 1 {
		return errors.New("config error: cm-cicada.airflow-server.timeout has invalid value")
	}

	if CMCicadaConfig.CMCicada.AirflowServer.Username == "" {
		return errors.New("config error: cm-cicada.airflow-server.username is empty")
	}
	if CMCicadaConfig.CMCicada.AirflowServer.Password == "" {
		return errors.New("config error: cm-cicada.airflow-server.password is empty")
	}

	if CMCicadaConfig.CMCicada.Listen.Port == "" {
		return errors.New("config error: cm-cicada.listen.port is empty")
	}
	port, err = strconv.Atoi(CMCicadaConfig.CMCicada.Listen.Port)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("config error: cm-cicada.listen.port has invalid value")
	}

	return nil
}

func readCMCicadaConfigFile() error {
	common.RootPath = os.Getenv(common.ModuleROOT)
	if len(common.RootPath) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		common.RootPath = homeDir + "/." + strings.ToLower(common.ModuleName)
	}

	err := fileutil.CreateDirIfNotExist(common.RootPath)
	if err != nil {
		return err
	}

	ex, err := os.Executable()
	if err != nil {
		return err
	}

	exPath := filepath.Dir(ex)
	configDir := exPath + "/conf"
	if !fileutil.IsExist(configDir) {
		configDir = common.RootPath + "/conf"
	}

	data, err := os.ReadFile(configDir + "/" + cmCicadaConfigFile)
	if err != nil {
		return errors.New("can't find the config file (" + cmCicadaConfigFile + ")" + fmt.Sprintln() +
			"Must be placed in '." + strings.ToLower(common.ModuleName) + "/conf' directory " +
			"under user's home directory or 'conf' directory where running the binary " +
			"or 'conf' directory where placed in the path of '" + common.ModuleROOT + "' environment variable")
	}

	err = yaml.Unmarshal(data, &CMCicadaConfig)
	if err != nil {
		return err
	}

	err = checkCMCicadaConfigFile()
	if err != nil {
		return err
	}

	return nil
}

func prepareCMCicadaConfig() error {
	err := readCMCicadaConfigFile()
	if err != nil {
		return err
	}

	return nil
}
