package git

import (
	"errors"
	"fmt"
	"log"

	git "github.com/libgit2/git2go/v31"
)

type GitRepo struct {
	repo      *git.Repository
	LocalPath string
	RemoteUrl string
}

// OpenGitRepo opens a git repository at localPath and switches to the branch. If the local repo
// does not exist the repo will be cloned from the remoteUrl.
func OpenGitRepo(remoteUrl, localPath, branch string) (*GitRepo, error) {

	var repo *git.Repository
	var err error

	repo, err = git.OpenRepository(localPath)
	if err != nil {
		log.Printf("%s not found, attempting to clone %s  (err = %v)", localPath, remoteUrl, err)
		repo, err = git.Clone(remoteUrl, localPath, &git.CloneOptions{})
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

// https://stackoverflow.com/questions/31496175/git2go-simulate-git-checkout-and-an-immediate-git-push?rq=1
// https://blog.gopheracademy.com/advent-2014/git2go-tutorial/

// get the git status of the repo
func (gitRepo *GitRepo) GitStatusAndCommit() error {

	opts := &git.StatusOptions{
		Flags: (git.StatusOptIncludeUntracked),
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
		if entry.Status == git.StatusWtNew {
			s := entry.IndexToWorkdir.NewFile.Path
			gitRepo.addToIndex(s)

		}
		// file is modified
		if entry.Status == git.StatusWtModified {
			s := entry.IndexToWorkdir.NewFile.Path
			gitRepo.addToIndex(s)

		}
		// fie us deleted
		if entry.Status == git.StatusWtDeleted {
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

	sig := &git.Signature{
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

func checkoutBranch(repo *git.Repository, branchName string) error {
	checkoutOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing | git.CheckoutAllowConflicts | git.CheckoutUseTheirs,
	}
	//Getting the reference for the remote branch
	// remoteBranch, err := repo.References.Lookup("refs/remotes/origin/" + branchName)
	remoteBranch, err := repo.LookupBranch("origin/"+branchName, git.BranchRemote)
	if err != nil {
		log.Print("Failed to find remote branch: " + branchName)
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

	localBranch, err := repo.LookupBranch(branchName, git.BranchLocal)
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
