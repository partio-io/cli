# Session Context

## User Prompts

### Prompt 1

# Minion Task: Implement the issue

You are a coding agent executing a task autonomously. Complete the task in a single session without human interaction.

## Issue

# Add Codex CLI agent integration

## Description

Extend Partio's agent detection to recognise OpenAI Codex CLI (`codex`) as a supported AI agent alongside the existing Claude Code integration.

Partio currently only detects Claude Code via process inspection (`internal/agent/claude/`). The `agent.Detector` interface is already ...

