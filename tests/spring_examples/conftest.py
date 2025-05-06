import os, pathlib, pytest, subprocess, shutil

ROOT = pathlib.Path(__file__).parents[2]
EXAMPLES = ROOT / "spring-boot-examples"

def _is_war(dir_: pathlib.Path) -> bool:
    pom = dir_ / "pom.xml"
    return pom.exists() and "<packaging>war" in pom.read_text()

def spring_dirs():
    for d in EXAMPLES.glob("spring-boot-*"):
        if d.is_dir() and not _is_war(d):
            yield d

@pytest.fixture(scope="session", autouse=True)
def _set_mock():
    os.environ.setdefault("OPENAI_MOCK", "1")   # offline, fast 