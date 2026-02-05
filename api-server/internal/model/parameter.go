package model

import "time"

// Parameter 参数信息
type Parameter struct {
	ID                uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	JobID             *string   `gorm:"column:job_id;index" json:"jobId"`
	ParameterRaw      *string   `gorm:"column:parameter_raw;type:longtext" json:"parameterRaw"`
	ParameterData     *string   `gorm:"column:parameter_data;type:json" json:"parameterData"`
	ParameterSource   *string   `gorm:"column:parameter_source" json:"parameterSource"`
	ConfigFilePath    *string   `gorm:"column:config_file_path;type:text" json:"configFilePath"`
	ConfigFileContent *string   `gorm:"column:config_file_content;type:longtext" json:"configFileContent"`
	EnvVars           *string   `gorm:"column:env_vars;type:json" json:"envVars"`
	Timestamp         time.Time `gorm:"column:timestamp" json:"timestamp"`
}

func (Parameter) TableName() string {
	return "parameters"
}
