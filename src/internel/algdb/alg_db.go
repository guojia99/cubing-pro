package algdb

type AlgDB interface {
	ID() []string
	// Cases 为空则代表无case， 只需要查询即可。
	Cases() []string
	/*
		Select
		selectInput:
		    csp 桶桶
		    bld edge ACE
			bld edge UF-UL-UB
		config:
			独立配置，可以为空
	*/
	UpdateCases() []string
	Help() string
	Select(selectInput string, config interface{}) (output string, err error)
	UpdateConfig(caseInput string, oldConfig interface{}) (config string, err error)
	BaseConfig() interface{}
}
