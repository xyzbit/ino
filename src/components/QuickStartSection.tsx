import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Copy, Check, Terminal, Download, Play, ExternalLink } from 'lucide-react';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { atomDark } from 'react-syntax-highlighter/dist/esm/styles/prism';

const QuickStartSection: React.FC = () => {
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);
  const [activeTab, setActiveTab] = useState(0);

  const codeExamples = [
    {
      title: 'Docker 部署',
      language: 'bash',
      code: `# 克隆项目
git clone https://github.com/your-org/kag-system.git
cd kag-system

# 使用 Docker Compose 启动
docker-compose up -d

# 查看服务状态
docker-compose ps`,
    },
    {
      title: 'API 调用',
      language: 'javascript',
      code: `// 初始化客户端
const client = new KAGClient({
  baseURL: 'http://localhost:8080',
  apiKey: 'your-api-key'
});

// 发送查询请求
const response = await client.query({
  question: '什么是知识图谱？',
  context: 'AI技术',
  maxTokens: 500
});

console.log(response.answer);`,
    },
    {
      title: 'Go SDK',
      language: 'go',
      code: `package main

import (
    "fmt"
    "github.com/your-org/kag-go-sdk"
)

func main() {
    client := kag.NewClient("your-api-key")
    
    response, err := client.Query(&kag.QueryRequest{
        Question: "解释RAG技术原理",
        Context:  "人工智能",
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(response.Answer)
}`,
    },
    {
      title: 'Python SDK',
      language: 'python',
      code: `from kag_client import KAGClient

# 初始化客户端
client = KAGClient(
    base_url="http://localhost:8080",
    api_key="your-api-key"
)

# 发送查询
response = client.query(
    question="什么是向量数据库？",
    context="数据库技术",
    max_tokens=500
)

print(response.answer)`,
    },
  ];

  const steps = [
    {
      number: '01',
      title: '环境准备',
      description: '安装 Docker 和 Docker Compose',
      icon: Download,
    },
    {
      number: '02',
      title: '项目部署',
      description: '克隆项目并启动服务',
      icon: Play,
    },
    {
      number: '03',
      title: 'API 调用',
      description: '使用 SDK 或直接调用 API',
      icon: Terminal,
    },
    {
      number: '04',
      title: '开始使用',
      description: '集成到您的应用程序中',
      icon: ExternalLink,
    },
  ];

  const copyToClipboard = async (text: string, index: number) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedIndex(index);
      setTimeout(() => setCopiedIndex(null), 2000);
    } catch (err) {
      console.error('复制失败:', err);
    }
  };

  return (
    <section className="py-20 bg-slate-900 relative overflow-hidden">
      {/* 背景装饰 */}
      <div className="absolute inset-0">
        <div className="absolute top-0 left-1/2 transform -translate-x-1/2 w-96 h-96 bg-cyan-500/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl" />
      </div>

      <div className="relative z-10 max-w-7xl mx-auto px-6">
        {/* 标题 */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-6">
            <span className="bg-gradient-to-r from-cyan-400 to-blue-500 bg-clip-text text-transparent">
              快速开始
            </span>
          </h2>
          <p className="text-xl text-gray-300 max-w-3xl mx-auto">
            几分钟内即可部署并开始使用 KAG 系统
          </p>
        </motion.div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
          {/* 步骤指南 */}
          <motion.div
            initial={{ opacity: 0, x: -50 }}
            whileInView={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.8 }}
            viewport={{ once: true }}
          >
            <h3 className="text-2xl font-bold text-white mb-8">部署步骤</h3>
            <div className="space-y-6">
              {steps.map((step, index) => (
                <motion.div
                  key={index}
                  initial={{ opacity: 0, y: 20 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.6, delay: index * 0.1 }}
                  viewport={{ once: true }}
                  className="flex items-start space-x-4 group cursor-pointer"
                >
                  <div className="flex-shrink-0">
                    <div className="w-12 h-12 bg-gradient-to-br from-cyan-500 to-blue-600 rounded-full flex items-center justify-center text-white font-bold text-sm group-hover:scale-110 transition-transform duration-300">
                      {step.number}
                    </div>
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-2">
                      <step.icon size={20} className="text-cyan-400" />
                      <h4 className="text-lg font-semibold text-white group-hover:text-cyan-400 transition-colors">
                        {step.title}
                      </h4>
                    </div>
                    <p className="text-gray-300">{step.description}</p>
                  </div>
                </motion.div>
              ))}
            </div>
          </motion.div>

          {/* 代码示例 */}
          <motion.div
            initial={{ opacity: 0, x: 50 }}
            whileInView={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.8 }}
            viewport={{ once: true }}
          >
            <h3 className="text-2xl font-bold text-white mb-8">代码示例</h3>
            
            {/* 标签页 */}
            <div className="flex space-x-2 mb-6 overflow-x-auto">
              {codeExamples.map((example, index) => (
                <button
                  key={index}
                  onClick={() => setActiveTab(index)}
                  className={`px-4 py-2 rounded-lg font-medium whitespace-nowrap transition-all duration-300 ${
                    activeTab === index
                      ? 'bg-cyan-500 text-white'
                      : 'bg-white/10 text-gray-300 hover:bg-white/20'
                  }`}
                >
                  {example.title}
                </button>
              ))}
            </div>

            {/* 代码块 */}
            <div className="relative">
              <div className="absolute top-4 right-4 z-10">
                <motion.button
                  whileHover={{ scale: 1.1 }}
                  whileTap={{ scale: 0.9 }}
                  onClick={() => copyToClipboard(codeExamples[activeTab].code, activeTab)}
                  className="p-2 bg-white/10 backdrop-blur-sm rounded-lg border border-white/20 text-white hover:bg-white/20 transition-all duration-300"
                >
                  {copiedIndex === activeTab ? (
                    <Check size={16} className="text-green-400" />
                  ) : (
                    <Copy size={16} />
                  )}
                </motion.button>
              </div>
              
              <div className="bg-gray-900 rounded-xl border border-gray-700 overflow-hidden">
                <div className="flex items-center space-x-2 px-4 py-3 bg-gray-800 border-b border-gray-700">
                  <div className="w-3 h-3 bg-red-500 rounded-full" />
                  <div className="w-3 h-3 bg-yellow-500 rounded-full" />
                  <div className="w-3 h-3 bg-green-500 rounded-full" />
                  <span className="ml-4 text-sm text-gray-400 font-mono">
                    {codeExamples[activeTab].title}
                  </span>
                </div>
                
                <SyntaxHighlighter
                  language={codeExamples[activeTab].language}
                  style={atomDark}
                  customStyle={{
                    margin: 0,
                    padding: '1.5rem',
                    background: 'transparent',
                    fontSize: '0.875rem',
                  }}
                >
                  {codeExamples[activeTab].code}
                </SyntaxHighlighter>
              </div>
            </div>
          </motion.div>
        </div>

        {/* 底部链接 */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.5 }}
          viewport={{ once: true }}
          className="text-center mt-16"
        >
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="px-8 py-4 bg-gradient-to-r from-cyan-500 to-blue-600 text-white font-semibold rounded-full shadow-lg hover:shadow-cyan-500/25 transition-all duration-300"
            >
              查看完整文档
            </motion.button>
            
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="px-8 py-4 border-2 border-cyan-400 text-cyan-400 font-semibold rounded-full hover:bg-cyan-400 hover:text-slate-900 transition-all duration-300"
            >
              下载 SDK
            </motion.button>
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default QuickStartSection;