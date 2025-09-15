# Autovibe

[![Discord](https://img.shields.io/discord/1415213817947750452?color=7289da&label=Discord&logo=discord&logoColor=white)](https://discord.gg/E6BVfQpT)

```bash
make vibe
```

Everything is an Intelligent Machine, even you.. jusk kidding (but it might be partially true).

So, everything is an Intelligent Machine and they speak with each other through the Hub, which is also an IM. Intelligent Machines essentially implement the https://en.wikipedia.org/wiki/Actor_model and overlap with [whitepaper.actor](whitepaper.actor). There are following main IM types:

## Hub

Hub orchestrates other IMs. It is a VM that has docker compose stack with microservices and other software needed to be able to create other machines, proxy traffic, etc etc.

## Agent

Agent IM runs one of many AI agents such as Claude Code, Gemini, etc. Agents that do other things like controlling a browser for example rather than producing software are totally valid.

## Project Coordinator

Project Coordinator (PC) is a specialized agent that evolves projects. Instead of creating an analogy of a Project Manager from real IT world, we believe it's better to allow abstraction and composition. PCs can run other PCs to delegate high level tasks such as researching libraries, coming up with architecture, doing marketing research, analyzing papers, etc, as well as running agents to accomplish coding, web navigating, research and other tasks.

## MVP

For the mvp, each IM is a virtual machine but post MVP we will use Unikernels to improve security and decrease RAM/disk usage.

## License

Currently, the license is yet to be drafted but the main idea stems from the future evolution plans of Web3 integration by ensuring each user either uses a paid version from AINicorn company or free one which has the Community Intelligent Machine that ensures a very small amount of CPU, RAM, disk space and network time is used to improve federated learning models powering Autovibe without disclosing any user's private data.