package start

import (
	"WgInspector/utils"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/17
 */

var localFileOptFunc = func(opt utils.Option) {}

func SetLocalConfigReaderOption(filePath, configType string) {
	localFileOptFunc = func(opt utils.Option) {
		opt["config_reader"] = "local_file"
		opt["config_parser"] = configType
		opt["filepath"] = filePath
	}
}

func WithConfigReader(driverName string) utils.OptionFunc {
	return func(opt utils.Option) {
		opt["config_reader"] = driverName
	}
}

func WithConfigParser(driverName string) utils.OptionFunc {
	return func(opt utils.Option) {
		opt["config_parser"] = driverName
	}
}
