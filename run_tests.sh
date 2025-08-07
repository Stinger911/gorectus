#!/bin/bash
set -ex
cd frontend
npm install
npm test -- --watchAll=false
