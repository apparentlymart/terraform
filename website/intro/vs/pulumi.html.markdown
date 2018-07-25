---
layout: "intro"
page_title: "Terraform vs. Pulumi"
sidebar_current: "vs-other-pulumi"
description: |-
  Pulumi is .
---

# Terraform vs. Pulumi

Pulumi is a tool for declaring and provisioning cloud infrastructure defined
using imperative programming languages, including JavaScript, TypeScript,
Python, and Go.

It is a close relative to Terraform in terms of its scope, but has some
philosophical differences that distinguish it.

## Interaction with Application Code

Pulumi has features that allow and encourage a tight integration between
application logic and infrastructure, defining both together as a single
program. A single source file may blend both the definition of an Amazon S3
bucket and some code to run in AWS Lambda when a new object is written into
it.

[The Tao of HashiCorp](https://www.hashicorp.com/blog/the-tao-of-hashicorp)
encourages us to create tools with a small and well-defined scope. Terraform
is focused on the problem of provisioning infrastructure, delegating tasks such
as building and testing application artifacts to other software that is tailored
to those problems.

Pulumi's design encourages implicit interactions between application code
and infrastructure, while Terraform instead encourages an explicit separation
of concerns. Infrastructure and application code are often changed on different
timescales, and different risk profiles for those changes.

## Imperative vs. Declarative Languages

Terraform uses a declarative domain-specific language to describe the intended
result. This language is interpreted by Terraform to determine the set of
requested resources and the relationships between them. Although Terraform
provides constructs such as expressions and modules to create simple
abstractions and factor out common configuration elements, the language is
designed so that it is easy to understand the correspondence between the
configuration constructs and the remote objects that they represent, and to
understand how proposed changes are likely to affect the infrastructure once
applied.

Pulumi instead builds this set of resources and their relationships by running
some imperative code you supply, written in one of a number of supported
languages. As a result, it allows more expressiveness and flexibility in how
configuration is defined, and so it allows the user to produce more
sophisticated abstractions that hide the details of exactly how a particular
system is implemented. As the complexity of these abstractions grows, it
may become more challenging to understand the correspondence between the
configuration and the real infrastructure.

Both approaches have advantages and disadvantages, and are suited to different
problems. Terraform's specialized language is easy to learn, and that learning
is readily transferrable between projects and teams. Pulumi's model allows for
richer user-defined abstractions, which can allow for more expressiveness in
definition but can therefore lead to significant differences in approach
between teams that have made different abstraction choices.

Because Terraform has a single language that is tightly integrated with it,
Terraform can more completely validate configuration and produce contextual
error messages in case of problems. When constructing configuration in an
imperative language as Pulumi allows, error messages will often be at the
abstraction level of the host language rather than of the infrastructure
being described. When Pulumi is used with dynamically-typed languages such as
JavaScript, consistency problems may be undetected during planning and become
apparent only after deployment.

Terraform can also support generated configuration via its alternative JSON
syntax, with a separate program generating JSON configuration files as a
pre-processing step. This allows both approaches to be blended to a certain
extent, although the extra pre-processing step can obscure the relationships
between the code and real resources. The Terraform team recommends against
extensive use of generated configuration for this reason.
