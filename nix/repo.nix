{ pkgs }: {
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
}
