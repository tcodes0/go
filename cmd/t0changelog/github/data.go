// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package github

var URLPrefix = "https://github.com/"

type FatCommit struct {
	SHA    string `json:"sha"`
	Commit Commit `json:"commit"`
}

type Commit struct {
	Message string `json:"message"`
}
