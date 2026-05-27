package gitrunnertest

import (
	"context"
	"sync"

	"git-clone-manager/internal/gitrunner"
)

type CloneCall struct {
	URL      string
	DestPath string
}

type Fake struct {
	mu sync.Mutex

	cloneFunc         func(url, destPath string) error
	originURLFunc     func(repoPath string) (string, error)
	fetchFunc         func(repoPath string) error
	currentBranchFunc func(repoPath string) (string, error)
	dirtyCountFunc    func(repoPath string) (int, error)
	commitsBehindFunc func(repoPath string) (int, error)
	defaultBranchFunc func(repoPath string) (string, error)

	currentBranchValue string
	originURLValue     string
	dirtyCountValue    int
	commitsBehindValue int
	defaultBranchValue string

	cloneCalls         []CloneCall
	originURLCalls     []string
	fetchCalls         []string
	currentBranchCalls []string
	dirtyCountCalls    []string
	commitsBehindCalls []string
	defaultBranchCalls []string
}

var _ gitrunner.Runner = (*Fake)(nil)

func New() *Fake {
	return &Fake{}
}

func (fakeRunner *Fake) StubClone(fn func(url, destPath string) error) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.cloneFunc = fn
}

func (fakeRunner *Fake) StubFetch(fn func(repoPath string) error) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.fetchFunc = fn
}

func (fakeRunner *Fake) StubOriginURL(fn func(repoPath string) (string, error)) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.originURLFunc = fn
}

func (fakeRunner *Fake) StubCurrentBranch(fn func(repoPath string) (string, error)) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.currentBranchFunc = fn
}

func (fakeRunner *Fake) StubDirtyCount(fn func(repoPath string) (int, error)) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.dirtyCountFunc = fn
}

func (fakeRunner *Fake) StubCommitsBehind(fn func(repoPath string) (int, error)) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.commitsBehindFunc = fn
}

func (fakeRunner *Fake) StubDefaultBranch(fn func(repoPath string) (string, error)) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.defaultBranchFunc = fn
}

func (fakeRunner *Fake) SetDirtyCount(value int) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.dirtyCountValue = value
}

func (fakeRunner *Fake) SetCurrentBranch(branch string) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.currentBranchValue = branch
}

func (fakeRunner *Fake) SetOriginURL(originURL string) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.originURLValue = originURL
}

func (fakeRunner *Fake) SetCommitsBehind(value int) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.commitsBehindValue = value
}

func (fakeRunner *Fake) SetDefaultBranch(branch string) {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	fakeRunner.defaultBranchValue = branch
}

func (fakeRunner *Fake) Clone(ctx context.Context, url, destPath string) error {
	fakeRunner.mu.Lock()
	fakeRunner.cloneCalls = append(fakeRunner.cloneCalls, CloneCall{URL: url, DestPath: destPath})
	cloneFunc := fakeRunner.cloneFunc
	fakeRunner.mu.Unlock()

	if cloneFunc != nil {
		return cloneFunc(url, destPath)
	}

	return nil
}

func (fakeRunner *Fake) OriginURL(ctx context.Context, repoPath string) (string, error) {
	fakeRunner.mu.Lock()
	fakeRunner.originURLCalls = append(fakeRunner.originURLCalls, repoPath)
	originURLFunc := fakeRunner.originURLFunc
	originURLValue := fakeRunner.originURLValue
	fakeRunner.mu.Unlock()

	if originURLFunc != nil {
		return originURLFunc(repoPath)
	}

	return originURLValue, nil
}

func (fakeRunner *Fake) Fetch(ctx context.Context, repoPath string) error {
	fakeRunner.mu.Lock()
	fakeRunner.fetchCalls = append(fakeRunner.fetchCalls, repoPath)
	fetchFunc := fakeRunner.fetchFunc
	fakeRunner.mu.Unlock()

	if fetchFunc != nil {
		return fetchFunc(repoPath)
	}

	return nil
}

func (fakeRunner *Fake) CurrentBranch(ctx context.Context, repoPath string) (string, error) {
	fakeRunner.mu.Lock()
	fakeRunner.currentBranchCalls = append(fakeRunner.currentBranchCalls, repoPath)
	currentBranchFunc := fakeRunner.currentBranchFunc
	currentBranchValue := fakeRunner.currentBranchValue
	fakeRunner.mu.Unlock()

	if currentBranchFunc != nil {
		return currentBranchFunc(repoPath)
	}

	return currentBranchValue, nil
}

func (fakeRunner *Fake) DirtyCount(ctx context.Context, repoPath string) (int, error) {
	fakeRunner.mu.Lock()
	fakeRunner.dirtyCountCalls = append(fakeRunner.dirtyCountCalls, repoPath)
	dirtyCountFunc := fakeRunner.dirtyCountFunc
	dirtyCountValue := fakeRunner.dirtyCountValue
	fakeRunner.mu.Unlock()

	if dirtyCountFunc != nil {
		return dirtyCountFunc(repoPath)
	}

	return dirtyCountValue, nil
}

func (fakeRunner *Fake) CommitsBehind(ctx context.Context, repoPath string) (int, error) {
	fakeRunner.mu.Lock()
	fakeRunner.commitsBehindCalls = append(fakeRunner.commitsBehindCalls, repoPath)
	commitsBehindFunc := fakeRunner.commitsBehindFunc
	commitsBehindValue := fakeRunner.commitsBehindValue
	fakeRunner.mu.Unlock()

	if commitsBehindFunc != nil {
		return commitsBehindFunc(repoPath)
	}

	return commitsBehindValue, nil
}

func (fakeRunner *Fake) DefaultBranch(ctx context.Context, repoPath string) (string, error) {
	fakeRunner.mu.Lock()
	fakeRunner.defaultBranchCalls = append(fakeRunner.defaultBranchCalls, repoPath)
	defaultBranchFunc := fakeRunner.defaultBranchFunc
	defaultBranchValue := fakeRunner.defaultBranchValue
	fakeRunner.mu.Unlock()

	if defaultBranchFunc != nil {
		return defaultBranchFunc(repoPath)
	}

	return defaultBranchValue, nil
}

func (fakeRunner *Fake) CloneCalls() []CloneCall {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	return append([]CloneCall(nil), fakeRunner.cloneCalls...)
}

func (fakeRunner *Fake) OriginURLCalls() []string {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	return append([]string(nil), fakeRunner.originURLCalls...)
}

func (fakeRunner *Fake) FetchCalls() []string {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	return append([]string(nil), fakeRunner.fetchCalls...)
}

func (fakeRunner *Fake) DirtyCountCalls() []string {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	return append([]string(nil), fakeRunner.dirtyCountCalls...)
}

func (fakeRunner *Fake) CurrentBranchCalls() []string {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	return append([]string(nil), fakeRunner.currentBranchCalls...)
}

func (fakeRunner *Fake) CommitsBehindCalls() []string {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	return append([]string(nil), fakeRunner.commitsBehindCalls...)
}

func (fakeRunner *Fake) DefaultBranchCalls() []string {
	fakeRunner.mu.Lock()
	defer fakeRunner.mu.Unlock()

	return append([]string(nil), fakeRunner.defaultBranchCalls...)
}
