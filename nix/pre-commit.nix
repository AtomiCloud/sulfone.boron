{ formatter, pre-commit-lib, packages }:
pre-commit-lib.run {
  src = ./.;

  # hooks
  hooks = {
    # formatter
    treefmt = {
      enable = true;
      excludes = [ ".*scaffold/templates/.*" ".*infra/.*" "pnpm-lock.yaml" "Changelog.md" "CommitConventions.md" ".*schema.json" ];
    };
    # linters From https://github.com/cachix/pre-commit-hooks.nix
    shellcheck = {
      enable = false;
    };

    a-infisical = {
      enable = true;
      name = "Secrets Scanning (Past Commits)";
      description = "Scan for possible secrets in past commits";
      entry = "${packages.infisical}/bin/infisical scan . -v";
      language = "system";
      pass_filenames = false;
    };

    a-infisical-staged = {
      enable = true;
      name = "Secrets Scanning (Staged)";
      description = "Scan for possible secrets in staged files";
      entry = "${packages.infisical}/bin/infisical scan git-changes --staged -v";
      language = "system";
      pass_filenames = false;
    };

    a-golang-ci-lint = {
      enable = true;
      name = "Golang CI Lint";
      description = "Lint Golang code";
      entry = "${packages.golangci-lint}/bin/golangci-lint run --timeout 10m0s .";
      language = "system";
      files = ".*go$";
      pass_filenames = false;
    };

    a-gitlint = {
      enable = true;
      name = "Gitlint";
      description = "Lints git commit message";
      entry = "${packages.gitlint}/bin/gitlint --staged --msg-filename .git/COMMIT_EDITMSG";
      language = "system";
      pass_filenames = false;
      stages = [ "commit-msg" ];
    };


    a-enforce-gitlint = {
      enable = true;
      name = "Enforce gitlint";
      description = "Enforce atomi_releaser conforms to gitlint";
      entry = "${packages.sg}/bin/sg gitlint";
      files = "(atomi_release\\.yaml|\\.gitlint)";
      language = "system";
      pass_filenames = false;
    };

    a-shellcheck = {
      enable = true;
      name = "Shell Check";
      entry = "${packages.shellcheck}/bin/shellcheck";
      files = ".*sh$";
      language = "system";
      pass_filenames = true;
    };

    a-enforce-exec = {
      enable = true;
      name = "Enforce Shell Script executable";
      entry = "${packages.coreutils}/bin/chmod +x";
      files = ".*sh$";
      language = "system";
      pass_filenames = true;
    };

    a-hadolint = {
      enable = true;
      name = "Docker Linter";
      entry = "${packages.hadolint}/bin/hadolint";
      files = ".*Dockerfile$";
      language = "system";
      pass_filenames = true;
    };


  };

  settings = {
    treefmt = {
      package = formatter;
    };
  };
}
