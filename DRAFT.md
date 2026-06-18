## Context

First of all, read the file specs/constitution.md and apply these rules throughout the your ongoing work.

read and understand how the marble testing feature works, can be currently used and how it is implemented by reading:
- all related documentation under the docs folder
- the source code directly under the eventest folder and subfolders

Important: don't code nothing for now, instead write a detailed technical specification in the new file TECHNICAL_SPEC.md.

Find a good and optimistic solution to these problems related to marble testing:

**the `initEvent` as the initial event**

- This event is the one passed by the user to the harness' RunAndWait method.
- This event MUST be unique
- I want to get rid of the existing naming and system (multiple placeholder/ sometimes called start event...)
- this event MUST ALWAYS be present in the expectation chain
- it MUST NOT be present in the side effect chain

**The test must be driven by the expectation**

- The test last all the expectation chain duration
- The side effect chain start at the same tick as the initEvent
- the side effect chain may be smaller than the expectation one, but never longer (an error must be risen if so)
- the method RunAndWait blocks until the end of the test
