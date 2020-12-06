# Developer Guide

This guide will help you get started developing Turing.

## Development Infrastructure

Turing requires 

```bash
kind create cluster --config=kind/config.yaml --wait=5m

istioctl install -f istio/minimal-operator.yaml
```