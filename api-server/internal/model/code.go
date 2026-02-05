package model

import "time"

// Code 代码信息
type Code struct {
	ID                uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	JobID             *string   `gorm:"column:job_id;index" json:"jobId"`
	ScriptPath        *string   `gorm:"column:script_path;type:text" json:"scriptPath"`
	ScriptContent     *string   `gorm:"column:script_content;type:longtext" json:"scriptContent"`
	ImportedLibraries *string   `gorm:"column:imported_libraries;type:text" json:"importedLibraries"`
	ConfigFiles       *string   `gorm:"column:config_files;type:text" json:"configFiles"`
	ShScriptPath      *string   `gorm:"column:sh_script_path;type:text" json:"shScriptPath"`
	ShScriptContent   *string   `gorm:"column:sh_script_content;type:longtext" json:"shScriptContent"`
	Timestamp         time.Time `gorm:"column:timestamp" json:"timestamp"`
}

func (Code) TableName() string {
	return "code"
}
