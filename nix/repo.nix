{ pkgs }: {
  shell = pkgs.mkShell rec {
    name = "squire";

    buildInputs = [
      pkgs.go_1_17
      pkgs.docker-compose
      pkgs.postgresql_13
      pkgs.pgquarrel
    ];
  };
}
