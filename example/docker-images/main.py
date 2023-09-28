#!/usr/bin/env python
from os import getenv
from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI()

@app.get("/", response_class=HTMLResponse)
def read_example():
    file = getenv("FILE", "none")
    ret = f"file: {file}\n".encode()
    if file != "none":
        ret += open(file, "rb").read()
    return ret


