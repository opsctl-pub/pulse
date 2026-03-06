# opsctl-pulse

Lightweight system metrics agent for [OpsCtl](https://github.com/opsctl-pub). Pushes heartbeat and resource utilization to OpsCtl over an outbound-only connection.

## What It Collects

CPU, memory, swap, disk, network, uptime, and optionally Docker container counts (running/total).

## How It Works

- Runs as a systemd service
- Pushes metrics every 60 seconds
- Outbound-only, works behind NAT and firewalls
