{ pkgs }: {
  shell = pkgs.mkShell rec {
    name = "squire";

    buildInputs = [
      pkgs.docker-compose
      pkgs.postgresql_13
      pkgs.pgquarrel
    ];
  };
}
