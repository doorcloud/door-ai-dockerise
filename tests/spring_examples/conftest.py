import os, pathlib, subprocess, pytest, shutil

EXAMPLES = pathlib.Path(__file__).parents[2] / "spring-boot-examples"

def _is_war(pom: pathlib.Path) -> bool:
    try:
        return "<packaging>war" in pom.read_text()
    except FileNotFoundError:
        return False

def spring_dirs():
    for d in EXAMPLES.glob("spring-boot-*"):
        if d.is_dir() and not _is_war(d / "pom.xml"):
            yield d

@pytest.fixture(autouse=True, scope="session")
def _set_mock():
    os.environ.setdefault("OPENAI_MOCK", "1") 