package lock

type Manager interface {
	DependencySignatures() map[string]string
	Signature() string
	IsLocked() bool
}

// LOCKING PROTOCOL:
// * Check to see if LOCKFILE is present for both the dependencies and the
//   current step before performing any operation on cached data. If it doesn't,
//   report that the data is not cached, and error out.
// * If they all exist, verify that the current step's dependency signatures
//   match the dependency signatures:
//     * Deserialize both the LOCKFILE for the current step, and the LOCKFILE
//       for all dependency steps. For example, if we're running an experiment,
//       we deserialize the LOCKFILE for the corpus and the configuration, as
//       well as the experiment.
//     * Check that the dependency signatures of the current step are the same
//       as the signatures contained in the dependency LOCKFILES. If they are
//       the same, proceed; if not, report they don't match, suggest how to
//       mitigate, and error out.
// * Delete current LOCKFILE. This is a safety measure.
// * Run the step; if the current run fails and we haven't overwritten any
//   files, re-write the old LOCKFILE; if we succeed, re-generate LOCKFILE and
//   write to disk.

// CorpusLock:
//     DependencySignatures is empty
//     Signature returns the SHA512 of every datafile inside the corpus

// SampleLock:
//     DependencySignatures is the SHA512 of the CorpusSignature
//     Signature is the SHA512 of every datafile in the sample, plus (perhaps) the name

// ConfigLock:
//     DependencySignatures contains the SHA512 of every datafile in the sample that powers it.
//     Signature contains the SHA512 of every datafile generated by the config steps

// ExperimentLock:
//     DependencySignatures contains the signatures of the configuration, as well as the sample we're using
//     Signature is empty?
