package apigithub

type FatCommit struct {
	SHA    string `json:"sha"`
	Commit Commit `json:"commit"`
}

type Commit struct {
	Message string `json:"message"`
}
