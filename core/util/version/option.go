package version

type option struct {
	major      uint8
	minor      uint8
	bugfix     uint8
	prerelease uint8
}

type Option func(op *option)

func WithMajor(major uint8) Option {
	return func(op *option) {
		op.major = major
	}
}

func WithMinor(minor uint8) Option {
	return func(op *option) {
		op.minor = minor
	}
}

func WithBugfix(bugfix uint8) Option {
	return func(op *option) {
		op.bugfix = bugfix
	}
}

func WithPrerelease(prerelease uint8) Option {
	return func(op *option) {
		op.prerelease = prerelease
	}
}
