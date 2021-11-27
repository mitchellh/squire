{ pkgs }: let
  src = ./..;

  # Get our version by reading the value in internal/version
  version = builtins.readFile(pkgs.runCommand "get-version" {} ''
    grep 'Version = "' ${src}/internal/version/version.go | \
      awk -F '( |\t)+' '{print $4}' | \
      tr -d '\n"' > $out
  '');
in rec {
  shell = pkgs.mkShell rec {
    name = "squire";

    buildInputs = [
      pkgs.cue
      pkgs.go_1_17
      pkgs.goreleaser
      pkgs.docker-compose
      pkgs.postgresql_14
      pkgs.pgquarrel
    ];
  };

  package = pkgs.buildGoModule {
    inherit version;
    name = "squire";
    src = ./..;
    subPackages = [ "cmd/squire" ];

    # This has to be updated each time the go.mod changes. Running a
    # nix build . should tell you this is wrong.
    vendorSha256 = "sha256-L0B44/E0p5FU6ylO/DhoXKsVfOEiUu7snedqc0KQCgc=";

    # There is no real reason to run the tests cause it only runs the
    # cmd/squire tests which do nothing except validate compilation. And
    # the full unit tests won't pass without devShell dependencies.
    doCheck = false;

    # We need wrapProgram
    nativeBuildInputs = [ pkgs.makeWrapper ];

    # Add dependent binaries to our PATH
    #
    # NOTE: we purposely prefix pgquarrel and suffix psql. It is more likely
    # that the user has a psql they want to use, so we prefer to use that.
    # However, pgquarrel we prefer to run our version.
    postInstall = ''
      wrapProgram "$out/bin/squire" \
        --prefix PATH : "${pkgs.pgquarrel}/bin" \
        --suffix PATH : "${pkgs.postgresql_14}/bin"
    '';
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
