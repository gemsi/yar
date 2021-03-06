package robber

import (
	"strings"
	"sync"
	"sync/atomic"

	"gopkg.in/src-d/go-git.v4/plumbing/transport"
)

const (
	// B64chars is used for entropy finding of base64 strings.
	B64chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	// Hexchars is used for entropy finding of hex based strings.
	Hexchars = "1234567890abcdefABCDEF"
	// Threshold for b64 matching of entropy strings
	b64Threshold = 4.5
	// Threshold for hex matching of entropy strings
	hexThreshold = 3
)

// AnalyzeEntropyDiff breaks a given diff into words and finds valid base64 and hex
// strings within a word and finally runs an entropy check on the valid string.
// Code taken from https://github.com/dxa4481/truffleHog.
func AnalyzeEntropyDiff(m *Middleware, diffObject *DiffObject) {
	words := strings.Fields(*diffObject.Diff)
	for _, word := range words {
		b64strings := FindValidStrings(word, B64chars)
		hexstrings := FindValidStrings(word, Hexchars)
		PrintEntropyFinding(b64strings, m, diffObject, b64Threshold)
		PrintEntropyFinding(hexstrings, m, diffObject, hexThreshold)
	}
}

// AnalyzeRegexDiff runs line by line on a given diff and runs each given regex rule on the line.
func AnalyzeRegexDiff(m *Middleware, diffObject *DiffObject) {
	lines := strings.Split(*diffObject.Diff, "\n")
	numOfLines := len(lines)

	for lineNum, line := range lines {
		for _, rule := range m.Rules {
			if found := rule.Regex.FindString(line); found != "" {
				start, end := Max(0, lineNum-*m.Flags.Context), Min(numOfLines, lineNum+*m.Flags.Context+1)
				context := lines[start:end]
				newDiff := strings.Join(context, "\n")
				secret := []int{strings.Index(newDiff, found)}
				secret = append(secret, secret[0]+len(found))

				secretString := newDiff[secret[0]:secret[1]]
				if *m.Flags.SkipDuplicates && !m.SecretExists(*diffObject.Reponame, secretString) {
					m.AddSecret(*diffObject.Reponame, secretString)
					finding := NewFinding(rule.Reason, secret, diffObject)
					m.Logger.LogFinding(finding, m, newDiff)
				} else if !*m.Flags.SkipDuplicates {
					finding := NewFinding(rule.Reason, secret, diffObject)
					m.Logger.LogFinding(finding, m, newDiff)
				}
			}
		}
	}
}

// AnalyzeRepo opens a given repository and extracts all diffs from it for later analysis.
func AnalyzeRepo(m *Middleware, id int, repoch <-chan string, quit chan<- bool, done <-chan bool, wg *sync.WaitGroup) {
	for {
		select {
		case reponame := <-repoch:
			repo, err := OpenRepo(m, reponame)
			if err != nil {
				if err == transport.ErrEmptyRemoteRepository {
					m.Logger.LogWarn("%s is empty\n", reponame)
					atomic.AddInt32(m.RepoCount, -1)
					if atomic.LoadInt32(m.RepoCount) == 0 {
						quit <- true
					}
					continue
				}
				m.Logger.LogFail("Unable to open repo %s: %s\n", reponame, err)
			}

			commits, err := GetCommits(m, repo, reponame)
			if err != nil {
				m.Logger.LogWarn("Unable to fetch commits for %s: %s\n", reponame, err)
				return
			}

			// Get all changes in correct order of commit history
			for index := range commits {
				commit := commits[len(commits)-index-1]
				changes, err := GetCommitChanges(commit)
				if err != nil {
					m.Logger.LogWarn("Unable to get commit changes for hash %s: %s\n", commit.Hash, err)
					continue
				}

				for _, change := range changes {
					diffs, filepath, err := GetDiffs(m, change, reponame)
					if err != nil {
						m.Logger.LogWarn("Unable to get diffs of %s: %s\n", change, err)
						continue
					}
					for _, diff := range diffs {
						diffObject := NewDiffObject(commit, &diff, &reponame, &filepath)
						if *m.Flags.Both {
							AnalyzeRegexDiff(m, diffObject)
							AnalyzeEntropyDiff(m, diffObject)
						} else if *m.Flags.Entropy {
							AnalyzeEntropyDiff(m, diffObject)
						} else {
							AnalyzeRegexDiff(m, diffObject)
						}
					}
				}
			}
			atomic.AddInt32(m.RepoCount, -1)
			if atomic.LoadInt32(m.RepoCount) == 0 {
				quit <- true
			}
		case <-done:
			wg.Done()
			return
		}
	}
}

// AnalyzeUser simply sends a GET request on githubs API for a given username
// and starts and analysis of each of the user's repositories.
func AnalyzeUser(m *Middleware, username string, repoch chan<- string) {
	repos := GetUserRepos(m, username)
	atomic.AddInt32(m.RepoCount, int32(len(repos)))
	for _, repo := range repos {
		repoch <- *repo
	}
}

// AnalyzeOrg simply sends two GET requests to githubs API, one for a given organizations
// repositories and one for its' members.
func AnalyzeOrg(m *Middleware, orgname string, repoch chan<- string) {
	var members []*string
	if *m.Flags.IncludeMembers {
		members = GetOrgMembers(m, orgname)
	} else {
		members = []*string{}
	}
	repos := GetOrgRepos(m, orgname)
	atomic.AddInt32(m.RepoCount, int32(len(repos)))

	for _, repo := range repos {
		repoch <- *repo
	}
	for _, member := range members {
		AnalyzeUser(m, *member, repoch)
	}
}
