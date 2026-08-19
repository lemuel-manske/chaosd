# Chaosd — Distributed Systems Failure Laboratory

## 1. Overview

Chaosd is a Go-based CLI tool for experimenting with real distributed systems running in Docker.

The user provides an existing `docker-compose.yml`. Chaosd starts and observes the system, then provides a controlled environment in which the user can deliberately manipulate communication and node availability.

The fundamental idea is:

> **Docker runs the distributed system; Chaosd controls the experiment around it.**

Chaosd is not intended to be another Docker wrapper, a microservices framework, or a generic application-testing tool. Its purpose is to make the failure behavior of distributed systems observable, controllable, reproducible, and eventually automatically discoverable.

---

## 2. Problem

Distributed systems are difficult to reason about because correctness depends not only on the behavior of individual processes, but also on communication between them.

Failures are particularly difficult because seemingly harmless events can interact:

- messages can be delayed;
- messages can be lost;
- messages can arrive out of order;
- messages can be duplicated;
- nodes can disappear;
- nodes can restart;
- network partitions can occur;
- failures can happen simultaneously or in sequence.

A distributed application may behave correctly under each individual failure while failing under a particular combination of failures.

Manually discovering such scenarios is difficult, and difficult-to-reproduce failures are especially problematic.

Chaosd aims to provide an environment where these situations can be deliberately created, recorded, reproduced, and eventually searched automatically.

---

## 3. Vision

The long-term vision is a laboratory for distributed systems where a user can take an existing distributed application and experiment with its failure behavior without modifying the application itself.

The conceptual workflow is:

~~~text
Docker Compose application
          |
          v
       Chaosd
          |
          +---- observe
          |
          +---- perturb
          |
          +---- record
          |
          +---- replay
          |
          +---- analyze
          |
          +---- eventually discover
                interesting failures
~~~

The system should make failure scenarios first-class objects rather than ad-hoc manual operations.

---

## 4. Core Concept

The central abstraction is an **experiment**.

An experiment consists of:

1. a distributed application;
2. an initial environment;
3. a sequence of controlled events;
4. observations produced during execution;
5. a resulting state/outcome;
6. enough information to reproduce the execution.

For example, conceptually:

~~~text
Experiment

Application:
    distributed-system.yml

Events:
    t=5s   partition node-a ↔ node-b
    t=10s  restart node-b
    t=20s  heal partition

Observations:
    logs
    network events
    container state
    application metrics

Result:
    success / failure / assertion violation
~~~

An experiment should be recordable and replayable.

---

## 5. Distributed System Model

Chaosd should initially treat the application as a black box.

The application may be:

- a distributed database;
- a message-processing system;
- a replicated service;
- a consensus implementation;
- a distributed cache;
- a storage system;
- an application composed of several cooperating services.

Chaosd should not initially require knowledge of the application's internal protocol.

It should understand the environment in terms of:

- containers;
- nodes;
- networks;
- communication paths;
- process lifecycle;
- observable events.

This keeps the laboratory independent from specific technologies or distributed algorithms.

---

## 6. Docker Integration

Docker is the initial execution environment.

The user supplies a normal Docker Compose configuration.

Chaosd should be able to:

1. load the Compose application;
2. start the application;
3. inspect its topology;
4. identify relevant containers/nodes;
5. observe their state;
6. manipulate communication between them;
7. manipulate node lifecycle;
8. collect experiment information;
9. stop and clean up the environment.

The distributed application should not need to be rewritten specifically for Chaosd.

The goal is to make Chaosd useful with both:

- applications specifically created for distributed-systems experiments;
- existing distributed applications that can be run through Docker Compose.

---

## 7. Network as the Primary Experimental Surface

The primary focus is network behavior.

Chaosd should be capable of deliberately introducing conditions such as:

- latency;
- packet loss;
- packet duplication;
- packet reordering;
- bandwidth restrictions;
- network partitions;
- temporary connectivity loss;
- recovery from previous conditions.

The intention is to reproduce realistic distributed-system communication failures rather than simply terminate processes.

Network manipulation should preferably operate at the Linux networking layer rather than requiring application-level proxies.

This allows the laboratory to affect the communication environment without requiring the distributed application to be aware of Chaosd.

---

## 8. Node Lifecycle as a Secondary Experimental Surface

Chaosd should also control container/node availability.

Relevant operations include:

- kill;
- pause;
- restart;
- stop;
- recover.

These operations can be combined with network failures.

For example:

~~~text
t=5s    partition A ↔ B
t=8s    restart B
t=12s   delay B ↔ C
t=20s   heal A ↔ B
~~~

The important capability is not the individual operation but the ability to **compose failures into controlled scenarios**.

---

## 9. Interactive Experiments

Chaosd should provide an interactive CLI for manually manipulating a running system.

Conceptually:

~~~text
chaosd <compose-file>
~~~

starts the laboratory around the specified distributed system.

The user can then inspect the running environment:

~~~text
chaosd ps
~~~

and perform experimental operations against selected nodes.

Conceptually:

~~~text
chaosd attach <node>
chaosd partition <node-a> <node-b>
chaosd delay <node-a> <node-b> <duration>
chaosd loss <node-a> <node-b> <percentage>
chaosd restart <node>
chaosd heal <node-a> <node-b>
~~~

The exact CLI syntax is intentionally not part of the current specification.

The important requirement is that the interface should make experimentation concise and composable.

---

## 10. Declarative Experiments

Interactive manipulation should be complemented by a declarative experiment representation.

An experiment should be expressible as a sequence of events:

~~~yaml
events:
  - at: 5s
    partition: [node-a, node-b]

  - at: 10s
    restart: node-b

  - at: 20s
    heal: [node-a, node-b]
~~~

The declarative representation enables:

- repeatability;
- version control;
- automation;
- experimentation at scale;
- deterministic scheduling;
- experiment comparison;
- replay.

The exact schema can evolve independently of the core concepts.

---

## 11. Reproducibility

Reproducibility is a first-class requirement.

Every experiment should produce a record containing enough information to understand and, where technically possible, reproduce the execution.

An experiment record should conceptually include:

~~~text
experiment definition
random seed
application configuration
topology
initial state
scheduled events
actual events
container state
network configuration
observations
result
~~~

The user should be able to reference an experiment and replay it.

Conceptually:

~~~text
chaosd replay <experiment>
~~~

A replay should attempt to recreate the same sequence of environmental events against an equivalent application environment.

This makes failures actionable rather than ephemeral.

---

## 12. Observation

Chaosd should not only manipulate the environment; it should record what happened.

The laboratory should maintain an experiment timeline containing events such as:

~~~text
experiment started
container started
network partition introduced
message impairment configured
container restarted
network partition removed
container stopped
experiment completed
~~~

Application-level information should initially remain optional.

Potential observations include:

- container state;
- Docker events;
- logs;
- network configuration;
- network statistics;
- resource information;
- application-provided metrics.

The goal is to correlate **environmental events with application behavior**.

---

## 13. Assertions and Correctness

Chaosd cannot assume that every distributed system has the same definition of correctness.

Different applications may care about different properties.

Examples include:

~~~text
a successful write must remain observable

replicas must eventually converge

a transaction must not disappear after commit

a request must not be processed twice

a service must remain available under a specified failure
~~~

Therefore, correctness should eventually be represented through **pluggable assertions** rather than being hard-coded into the laboratory.

Possible assertion mechanisms include:

- HTTP checks;
- command execution;
- application-specific scripts;
- metric conditions;
- custom observers.

The initial laboratory does not need an elaborate assertion framework, but the architecture should leave room for it.

---

## 14. Application Awareness

Chaosd should initially remain application-agnostic.

The core laboratory should not need to understand:

- Raft;
- Paxos;
- Kafka semantics;
- database internals;
- HTTP semantics;
- CRDTs;
- consensus protocols.

Instead, application-specific knowledge should be introduced through optional observers or adapters.

This preserves the distinction between:

~~~text
Chaosd:
    controls and observes the environment

Application:
    implements its own distributed behavior
~~~

---

## 15. Failure Scenarios

The laboratory should support combinations of events rather than only isolated failures.

A useful scenario might involve:

~~~text
normal operation
        |
        v
network partition
        |
        v
node restart
        |
        v
message delay
        |
        v
partition recovery
        |
        v
application observation
~~~

The interesting behavior often emerges from the **interaction between events**, not from individual faults.

Therefore, event sequences are more important than individual fault types.

---

## 16. Automated Experimentation

The long-term direction is to allow Chaosd to automatically generate failure scenarios.

Instead of:

~~~text
human
  |
  +--> partition
  +--> restart
  +--> delay
~~~

the system should eventually be capable of:

~~~text
                 experiment space
                       |
                       v
                generate scenario
                       |
                       v
                execute system
                       |
                       v
                evaluate result
                    /       \
                   /         \
              normal       violation
                |              |
              discard        record
                               |
                               v
                           reproduce
~~~

The system should search for combinations of failures that produce interesting or incorrect behavior.

---

## 17. Central Research Question

The project should ultimately investigate a question such as:

> **How effectively can controlled and systematic failure exploration discover problematic executions in distributed systems compared with conventional randomized fault injection?**

This creates an experimental research direction rather than making Chaosd merely a tooling project.

Possible comparisons include:

- random versus systematic failure generation;
- unguided versus guided exploration;
- large failure sequences versus minimized sequences;
- different strategies for prioritizing experiments.

The goal is not to invent a new distributed protocol.

The goal is to study how distributed systems can be **experimentally challenged and analyzed under complex failure conditions**.

---

## 18. Failure Minimization

A particularly valuable extension is minimizing discovered failures.

Suppose an experiment fails after:

~~~text
partition A-B
delay B-C
restart C
loss A-D
restart B
heal A-B
delay C-D
~~~

but only three events are actually necessary.

Chaosd should eventually attempt to reduce the scenario to something like:

~~~text
partition A-B
restart C
delay B-C
~~~

The resulting artifact is substantially easier for a developer or researcher to understand and reproduce.

This creates another useful research question:

> **Can failure scenarios be automatically reduced to small, human-understandable counterexamples?**

---

## 19. Scope Philosophy

Chaosd should prioritize **depth over breadth**.

The project should not attempt to become:

- a complete chaos-engineering platform;
- a Kubernetes replacement;
- a general-purpose monitoring platform;
- a distributed database;
- a service mesh;
- a production orchestration system.

Its purpose is narrower:

> **Provide a controlled, reproducible environment for studying how distributed systems behave when their execution environment fails.**

Docker is the practical boundary that makes the project achievable while still allowing experiments against real software.

---

## 20. Final Product

The finished project should allow a user to move through the following workflow:

~~~text
1. Provide Docker Compose application
              |
              v
2. Start distributed system
              |
              v
3. Inspect topology and state
              |
              v
4. Introduce controlled failures
              |
              v
5. Observe application behavior
              |
              v
6. Record the experiment
              |
              v
7. Replay the experiment
              |
              v
8. Define correctness conditions
              |
              v
9. Automatically search for failures
              |
              v
10. Minimize and reproduce interesting failures
~~~

The fundamental value proposition is:

> **Take a real distributed system, put it in a controlled laboratory, deliberately make its world unreliable, and turn the resulting behavior into reproducible experiments.**

The project therefore sits at the intersection of **distributed systems, operating-system networking, fault injection, testing, observability, and experimental methodology**.
