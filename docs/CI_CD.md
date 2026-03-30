# CI/CD Integrations

## GitHub Actions

Install `aso` using the setup action:

```yaml
- uses: ASOManiac/setup-aso@v1
  with:
    version: latest

- run: aso --help
```

For end-to-end examples, see:
https://github.com/ASOManiac/setup-aso

## GitLab CI/CD Components

Use the `aso-ci-components` repository:

```yaml
include:
  - component: gitlab.com/ASOManiac/aso-ci-components/run@main
    inputs:
      stage: deploy
      job_prefix: release
      aso_version: latest
      command: aso --help
```

For install/run templates and self-managed examples:
https://github.com/ASOManiac/aso-ci-components

## Bitrise

Use the `setup-aso` Bitrise step repository:

```yaml
workflows:
  primary:
    steps:
    - git::https://github.com/ASOManiac/steps-setup-aso.git@main:
        inputs:
        - mode: run
        - version: latest
        - command: aso --help
```

## CircleCI

Use the CircleCI orb repository:
https://github.com/ASOManiac/aso-orb
