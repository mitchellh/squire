{ pkgs }: let
  src = ./..;
in rec {
  shell = pkgs.mkShell rec {
    name = "squire";

    buildInputs = [
      pkgs.cue
      pkgs.go_1_17
      pkgs.goreleaser
      pkgs.docker-compose
      pkgs.postgresql_13
      pkgs.pgquarrel
    ];
  };

  package = pkgs.buildGoModule {
    name = "squire";
    src = ./..;
    subPackages = [ "cmd/squire" ];

    # This has to be updated each time the go.mod changes. Running a
    # nix build . should tell you this is wrong.
    vendorSha256 = "sha256-IxzjrTM9XFCFGyqGzy1IpIdGnA7elU4FzIcrC+G/l5c=";
  };

  # fmtcheck verifies that our Go files are all formatted.
  fmtcheck = pkgs.runCommand "fmtcheck"
    {
      buildInputs = shell.buildInputs;
    }
    ''
      mkdir $out
      test -z $(gofmt -l ${src})
    '';
}
