package repo

import (
	"log"
	"os"

	"gopkg.in/libgit2/git2go.v24"

	"github.com/PrairieLearn/autograd/grader"
)

func Sync(repoURL, commit, autogradRoot, publicKey, privateKey, passphrase string) error {
	path := grader.GetGraderRoot(autogradRoot)

	log.Printf("Syncing grader repo %s", repoURL)

	callbacks := git.RemoteCallbacks{
		CertificateCheckCallback: makeCertificateCheckCallback(),
		CredentialsCallback:      makeCredentialsCallback(publicKey, privateKey, passphrase),
	}
	checkoutOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutForce,
	}
	fetchOpts := &git.FetchOptions{
		RemoteCallbacks: callbacks,
		DownloadTags:    git.DownloadTagsAll,
	}
	cloneOpts := &git.CloneOptions{
		CheckoutOpts: checkoutOpts,
		FetchOptions: fetchOpts,
	}

	repo, err := initializeRepo(repoURL, path, cloneOpts)
	if err != nil {
		return err
	}

	log.Println("Fetching remote origin")
	err = fetchOrigin(repo, fetchOpts)

	if isSHA1Hash(commit) {
		oid, err := git.NewOid(commit)
		if err != nil {
			return err
		}
		if err := repo.SetHeadDetached(oid); err != nil {
			return err
		}
	} else {
		if err := repo.SetHead(commit); err != nil {
			return err
		}
	}

	log.Printf("Checking out commit/ref '%s'", commit)
	if err := repo.CheckoutHead(checkoutOpts); err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}
	log.Printf("Repo sync success, HEAD at %s", head.Target().String())

	return nil
}

func initializeRepo(repoURL, path string, cloneOpts *git.CloneOptions) (*git.Repository, error) {
	shouldClone := false

	repo, err := git.OpenRepository(path)
	if err != nil {
		shouldClone = true
	} else {
		remote, err := repo.Remotes.Lookup("origin")
		if err != nil || remote.Url() != repoURL {
			err := os.RemoveAll(path)
			if err != nil {
				return nil, err
			}
			shouldClone = true
		}
	}

	if shouldClone {
		repo, err = git.Clone(repoURL, path, cloneOpts)
		if err != nil {
			return nil, err
		}
	}

	return repo, nil
}

func fetchOrigin(repo *git.Repository, fetchOpts *git.FetchOptions) error {
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		return err
	}
	if err := remote.Fetch(nil, fetchOpts, ""); err != nil {
		return err
	}
	return nil
}

func makeCertificateCheckCallback() git.CertificateCheckCallback {
	return func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
		return git.ErrOk
	}
}

func makeCredentialsCallback(publicKey, privateKey, passphrase string) git.CredentialsCallback {
	return func(url string, username_from_url string, allowed_types git.CredType) (git.ErrorCode, *git.Cred) {
		errCode, cred := git.NewCredSshKey(username_from_url, publicKey, privateKey, passphrase)
		return git.ErrorCode(errCode), &cred
	}
}

func isSHA1Hash(s string) bool {
	return len(s) == 40
}
