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

	if err = checkoutBranch(repo, branch); err != nil {
		log.Fatalf("Failed to checkout branch %s: %v", branch, err)
	}
	return &GitRepo{repo, localPath, remoteUrl}, nil
}

// https://stackoverflow.com/questions/31496175/git2go-simulate-git-checkout-and-an-immediate-git-push?rq=1
// https://blog.gopheracademy.com/advent-2014/git2go-tutorial/

// get the git status of the repo
func (gitRepo *GitRepo) GitStatus() (string, error) {
	list, err := gitRepo.repo.StatusList(&git.StatusOptions{
		Flags: git.StatusOptIncludeUntracked | git.StatusOptIncludeIgnored,
	})
	if err != nil {
		log.Printf("Failed to get status: %v", err)
		return "", err
	}

	count, _ := list.EntryCount()
	s := fmt.Sprintf("%v  count=%d", list, count)
	log.Println(s)

	for i := 0; i < count; i++ {
		entry, _ := list.ByIndex(i)
		fmt.Printf("%+v  status=0x%x\n", entry, entry.Status)
		// file is newly added
		if entry.Status == git.StatusWtNew {
			s := entry.IndexToWorkdir.NewFile.Path
			fmt.Printf("Its new %+v  newfile=%s\n", entry.IndexToWorkdir, s)

		}
		// file is modified
		if entry.Status == git.StatusWtModified {
			s := entry.IndexToWorkdir.NewFile.Path
			fmt.Printf("Its modified %+v  newfile=%s\n", entry.IndexToWorkdir, s)
		}

		if entry.Status == git.StatusIndexNew {
			fmt.Println("new")
		}

	}

	gitRepo.repo.DiffIndexToWorkdir(nil, &git.DiffOptions{})

	return s, nil
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
