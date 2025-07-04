大模型应用当前有着 没有长期记忆、缺乏业务知识、缺乏实时性等问题，这些问题的本质是大模型自主获取｜记录部分知识，比如：
- 企业私有知识
- 记忆数据，如用户行为、反馈、偏好...
这些可以统称为知识，当前有着 RAG、Memeroy System等等解决方案，去补充各种类型的知识，但是接入复杂，有很多重复工作。所以有了 KAG（Knowledge-Augmented Generation）或者叫 统一检索框架（Unified Retrieval Framework），这个系统的核心思想是：将所有可能对LLM有帮助的信息源（外部文档、对话历史、用户画像等）视为可检索的“知识”，并建立一个统一的框架来智能地检索、筛选、整合这些知识，最后以最优化的方式注入到Prompt中。


├── bin
├── cmd
├── config
├── deploy
├── docs
├── internal
│   ├── application
│   │   ├── collector
│   │   ├── evaluation
│   │   └── retriever
│   ├── infra
│   │   ├── mysql
│   │   └── redis
│   ├── repo
│   └── server
└── pkg