package oodlehttp

import "errors"

// ErrNotFound reports that an object no longer exists.
//
// Resources translate it into "removed from state" rather than an
// error, so that an object deleted outside Terraform is recreated
// instead of wedging every subsequent plan.
var ErrNotFound = errors.New("not found")
