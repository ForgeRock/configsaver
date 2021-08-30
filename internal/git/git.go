/*
 *
 * Copyright  2021 ForgeRock AS
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package git

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	g "github.com/libgit2/git2go/v31"
)

const (
	defaultGitRepo = "https://stash.forgerock.org/scm/cloud/forgeops.git"
)

type GitRepo struct {
	repo      *g.Repository
	LocalPath string
	RemoteUrl string
}

// OpenGitRepo opens a git repository at localPath and switches to the branch. If the local repo
// does not exist the repo will be cloned from the remoteUrl.
func OpenGitRepo(localPath, branch string) (*GitRepo, error) {

	var repo *g.Repository
	var err error

	remoteUrl := os.Getenv("GIT_REPO")
	if remoteUrl == "" {
		remoteUrl = defaultGitRepo
		log.Printf("GIT_REPO env var not provided. defaulting to %s\n", remoteUrl)
	}

	repo, err = g.OpenRepository(localPath)
	if err != nil {
		log.Printf("%s not found, attempting to clone %s", localPath, remoteUrl)
		cloneOptions := &g.CloneOptions{}
		sshPath := os.Getenv("GIT_SSH_PATH")
		if sshPath != "" {
			fmt.Printf("Configuring ssh credentials\n")
			if _, err := os.Stat(sshPath); err != nil {
				log.Fatalf("GIT_SSH_PATH path %s does not exist or is not readable", sshPath)
			}
			cloneOptions = &g.CloneOptions{
				FetchOptions: &g.FetchOptions{
					RemoteCallbacks: g.RemoteCallbacks{
						CredentialsCallback:      credentialsCallback,
						CertificateCheckCallback: certificateCheckCallback,
					},
				},
			}
		}
		repo, err = g.Clone(remoteUrl, localPath, cloneOptions)
		if err != nil {
			log.Fatal(err)
		}
	}

	// This will refresh the working tree with the current branch
	// This is probably what we want most of the time. Files deleted in the working directory
	// get restored
	if err = checkoutBranch(repo, branch); err != nil {
		log.Fatalf("Failed to checkout branch %s: %v", branch, err)
	}
	return &GitRepo{repo, localPath, remoteUrl}, nil
}

func credentialsCallback(urlstring, username string, allowedTypes g.CredType) (*g.Cred, error) {
	sshPath := os.Getenv("GIT_SSH_PATH")
	log.Printf("ssh credential callback path: %s", sshPath)
	cred, err := g.NewCredSshKey("git", filepath.Join(sshPath, "id_rsa.pub"), filepath.Join(sshPath, "id_rsa"), "")
	log.Printf("Credentials %v err %v", cred, err)
	return cred, err
}

// needed just for testing
func certificateCheckCallback(cert *g.Certificate, valid bool, hostname string) g.ErrorCode {
	log.Printf("Cert callback")
	return 0
}

// https://stackoverflow.com/questions/31496175/git2go-simulate-git-checkout-and-an-immediate-git-push?rq=1
// https://blog.gopheracademy.com/advent-2014/git2go-tutorial/

// get the git status of the repo, commit any changed files.
func (gitRepo *GitRepo) GitStatusAndCommit() error {

	opts := &g.StatusOptions{
		Flags: (g.StatusOptIncludeUntracked),
	}
	list, err := gitRepo.repo.StatusList(opts)

	if err != nil {
		log.Printf("Failed to get status: %v", err)
		return err
	}

	count, _ := list.EntryCount()
	log.Printf("Processing %d git changes\n", count)

	for i := 0; i < count; i++ {
		entry, _ := list.ByIndex(i)
		fmt.Printf("%+v\nstatus=0x%x\n\n", entry, entry.Status)
		// file is newly added
		if entry.Status == g.StatusWtNew {
			s := entry.IndexToWorkdir.NewFile.Path
			gitRepo.addToIndex(s)

		}
		// file is modified
		if entry.Status == g.StatusWtModified {
			s := entry.IndexToWorkdir.NewFile.Path
			gitRepo.addToIndex(s)

		}
		// fie us deleted
		if entry.Status == g.StatusWtDeleted {
			s := entry.IndexToWorkdir.NewFile.Path
			fmt.Printf("File removed %s\n", s)
			gitRepo.removeFromIndex(s)
		}
	}
	if count > 0 {
		gitRepo.Commit("automated commit")
	}

	return nil
}

// See https://github.com/libgit2/libgit2/blob/091165c53b2bcd5d41fb71d43ed5a23a3d96bf5d/tests/object/commit/commitstagedfile.c#L21-L134
// add a file to the index. Equivalent to git add file
func (gitRepo *GitRepo) addToIndex(path string) error {
	log.Printf("Adding %s to index\n", path)
	index, err := gitRepo.repo.Index()
	checkErr(err)
	err = index.AddByPath(path)
	checkErr(err)
	err = index.Write()
	checkErr(err)
	return nil
}

// remove a file from the index. Equivalent to git rm file
func (gitRepo *GitRepo) removeFromIndex(path string) error {
	log.Printf("removing %s from index\n", path)
	index, err := gitRepo.repo.Index()
	checkErr(err)
	err = index.RemoveByPath(path)
	checkErr(err)
	err = index.Write()
	checkErr(err)
	return nil
}

// Commit current index to the repo
func (gitRepo *GitRepo) Commit(message string) error {

	sig := &g.Signature{
		Name:  "config-saver",
		Email: "config-saver@forgerock.com",
	}
	index, err := gitRepo.repo.Index()
	checkErr(err)
	treeId, err := index.WriteTree()
	checkErr(err)
	tree, err := gitRepo.repo.LookupTree(treeId)
	checkErr(err)
	err = index.Write()
	checkErr(err)
	currentBranch, err := gitRepo.repo.Head()
	checkErr(err)

	currentTip, err := gitRepo.repo.LookupCommit(currentBranch.Target())
	checkErr(err)

	_, err = gitRepo.repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
	checkErr(err)
	return err
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// From https://gist.github.com/danielfbm/ba4ae91efa96bb4771351bdbd2c8b06f

func checkoutBranch(repo *g.Repository, branchName string) error {
	checkoutOpts := &g.CheckoutOpts{
		Strategy: g.CheckoutSafe | g.CheckoutRecreateMissing | g.CheckoutAllowConflicts | g.CheckoutUseTheirs,
		ProgressCallback: func(path string, completed, total uint) g.ErrorCode {
			log.Printf("cloning %s completed %d of %d\n", path, completed, total)
			return 0
		},
	}
	//Getting the reference for the remote branch
	// remoteBranch, err := repo.References.Lookup("refs/remotes/origin/" + branchName)
	remoteBranch, err := repo.LookupBranch("origin/"+branchName, g.BranchRemote)
	if err != nil {
		log.Print("Failed to find remote branch: " + branchName)
		// TODO: This fails if the remote branch does not exist (examp]e: origin/autosave)
		// Instead of generating an error here we should attempt to create the remote branch
		// git push --set-uptream-to=origin/autosave
		return err
	}
	defer remoteBranch.Free()

	// Lookup for commit from remote branch
	commit, err := repo.LookupCommit(remoteBranch.Target())
	if err != nil {
		log.Print("Failed to find remote branch commit: " + branchName)
		return err
	}
	defer commit.Free()

	localBranch, err := repo.LookupBranch(branchName, g.BranchLocal)
	// No local branch, lets create one
	if localBranch == nil || err != nil {
		// Creating local branch
		localBranch, err = repo.CreateBranch(branchName, commit, false)
		if err != nil {
			log.Print("Failed to create local branch: " + branchName)
			return err
		}

		// Setting upstream to origin branch
		err = localBranch.SetUpstream("origin/" + branchName)
		if err != nil {
			log.Print("Failed to create upstream to origin/" + branchName)
			return err
		}
	}
	if localBranch == nil {
		return errors.New("error while locating/creating local branch")
	}
	defer localBranch.Free()

	// Getting the tree for the branch
	localCommit, err := repo.LookupCommit(localBranch.Target())
	if err != nil {
		log.Print("Failed to lookup for commit in local branch " + branchName)
		return err
	}
	defer localCommit.Free()

	tree, err := repo.LookupTree(localCommit.TreeId())
	if err != nil {
		log.Print("Failed to lookup for tree " + branchName)
		return err
	}
	defer tree.Free()

	// Checkout the tree
	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		log.Print("Failed to checkout tree " + branchName)
		return err
	}
	// Setting the Head to point to our branch
	repo.SetHead("refs/heads/" + branchName)
	return nil
}
