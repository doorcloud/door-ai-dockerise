import subprocess, pathlib, pytest, textwrap

from conftest import spring_dirs

@pytest.mark.parametrize("example", list(spring_dirs()))
def test_spring_example_builds(example, tmp_path):
    log = tmp_path / f"{example.name}.log"
    cmd = ["go", "run", "./cmd/dockergen", str(example)]
    result = subprocess.run(cmd, stdout=log.open("w"),
                            stderr=subprocess.STDOUT)
    assert result.returncode == 0, textwrap.dedent(
        f"❌ {example.name} failed\n--- LOG ---\n{log.read_text()}"
    ) 