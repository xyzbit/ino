import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Code, Database, Cloud, Cpu, Network, Lock } from 'lucide-react';

const TechStackSection: React.FC = () => {
  const [visibleBars, setVisibleBars] = useState(false);

  const techCategories = [
    {
      icon: Code,
      title: '前端技术',
      color: 'from-blue-500 to-cyan-500',
      skills: [
        { name: 'React', level: 95 },
        { name: 'TypeScript', level: 90 },
        { name: 'Tailwind CSS', level: 88 },
        { name: 'Framer Motion', level: 85 },
      ],
    },
    {
      icon: Database,
      title: '后端技术',
      color: 'from-green-500 to-emerald-500',
      skills: [
        { name: 'Go', level: 92 },
        { name: 'MySQL', level: 88 },
        { name: 'Redis', level: 85 },
        { name: 'Neo4j', level: 80 },
      ],
    },
    {
      icon: Cpu,
      title: 'AI技术',
      color: 'from-purple-500 to-pink-500',
      skills: [
        { name: 'LLM', level: 90 },
        { name: 'Vector DB', level: 88 },
        { name: 'RAG', level: 85 },
        { name: 'Knowledge Graph', level: 82 },
      ],
    },
    {
      icon: Cloud,
      title: '云原生',
      color: 'from-orange-500 to-red-500',
      skills: [
        { name: 'Docker', level: 90 },
        { name: 'Kubernetes', level: 85 },
        { name: 'Microservices', level: 88 },
        { name: 'API Gateway', level: 80 },
      ],
    },
  ];

  const techTags = [
    'Go', 'React', 'TypeScript', 'MySQL', 'Redis', 'Neo4j', 'Milvus',
    'Docker', 'Kubernetes', 'LLM', 'RAG', 'Vector Database', 'Knowledge Graph',
    'Microservices', 'API Gateway', 'Tailwind CSS', 'Framer Motion', 'Vite'
  ];

  useEffect(() => {
    const timer = setTimeout(() => setVisibleBars(true), 500);
    return () => clearTimeout(timer);
  }, []);

  return (
    <section className="py-20 bg-gradient-to-br from-slate-800 to-slate-900 relative overflow-hidden">
      {/* 六边形背景装饰 */}
      <div className="absolute inset-0">
        {Array.from({ length: 20 }).map((_, i) => (
          <motion.div
            key={i}
            className="absolute w-16 h-16 border border-cyan-500/20"
            style={{
              left: `${Math.random() * 100}%`,
              top: `${Math.random() * 100}%`,
              clipPath: 'polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)',
            }}
            animate={{
              rotate: [0, 360],
              scale: [1, 1.2, 1],
            }}
            transition={{
              duration: 20 + Math.random() * 10,
              repeat: Infinity,
              ease: 'linear',
            }}
          />
        ))}
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
            <span className="bg-gradient-to-r from-cyan-400 to-purple-500 bg-clip-text text-transparent">
              技术栈
            </span>
          </h2>
          <p className="text-xl text-gray-300 max-w-3xl mx-auto">
            采用现代化技术栈，构建高性能、可扩展的智能系统
          </p>
        </motion.div>

        {/* 技术分类 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-16">
          {techCategories.map((category, categoryIndex) => (
            <motion.div
              key={categoryIndex}
              initial={{ opacity: 0, x: categoryIndex % 2 === 0 ? -50 : 50 }}
              whileInView={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.8, delay: categoryIndex * 0.2 }}
              viewport={{ once: true }}
              className="bg-white/5 backdrop-blur-sm rounded-2xl p-8 border border-white/10 hover:border-white/20 transition-all duration-300"
            >
              {/* 分类标题 */}
              <div className="flex items-center space-x-4 mb-6">
                <div className={`p-3 rounded-xl bg-gradient-to-br ${category.color}`}>
                  <category.icon size={24} className="text-white" />
                </div>
                <h3 className="text-2xl font-bold text-white">{category.title}</h3>
              </div>

              {/* 技能进度条 */}
              <div className="space-y-4">
                {category.skills.map((skill, skillIndex) => (
                  <div key={skillIndex}>
                    <div className="flex justify-between items-center mb-2">
                      <span className="text-gray-300 font-medium">{skill.name}</span>
                      <span className="text-cyan-400 font-bold">{skill.level}%</span>
                    </div>
                    <div className="w-full bg-gray-700 rounded-full h-2 overflow-hidden">
                      <motion.div
                        className={`h-full bg-gradient-to-r ${category.color} rounded-full`}
                        initial={{ width: 0 }}
                        animate={visibleBars ? { width: `${skill.level}%` } : { width: 0 }}
                        transition={{ duration: 1, delay: categoryIndex * 0.2 + skillIndex * 0.1 }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </motion.div>
          ))}
        </div>

        {/* 技术标签云 */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.5 }}
          viewport={{ once: true }}
          className="text-center"
        >
          <h3 className="text-2xl font-bold text-white mb-8">技术标签</h3>
          <div className="flex flex-wrap justify-center gap-4">
            {techTags.map((tag, index) => (
              <motion.span
                key={index}
                initial={{ opacity: 0, scale: 0 }}
                whileInView={{ opacity: 1, scale: 1 }}
                transition={{ duration: 0.5, delay: index * 0.05 }}
                viewport={{ once: true }}
                whileHover={{ scale: 1.1, y: -5 }}
                className="px-4 py-2 bg-gradient-to-r from-cyan-500/20 to-blue-500/20 border border-cyan-500/30 rounded-full text-cyan-300 font-medium hover:from-cyan-500/30 hover:to-blue-500/30 hover:border-cyan-400/50 transition-all duration-300 cursor-pointer"
              >
                {tag}
              </motion.span>
            ))}
          </div>
        </motion.div>

        {/* 架构图标 */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.8 }}
          viewport={{ once: true }}
          className="mt-16 text-center"
        >
          <div className="flex justify-center items-center space-x-8">
            {[
              { icon: Network, label: '微服务架构', color: 'text-blue-400' },
              { icon: Lock, label: '安全防护', color: 'text-green-400' },
              { icon: Cloud, label: '云原生部署', color: 'text-purple-400' },
            ].map((item, index) => (
              <motion.div
                key={index}
                whileHover={{ scale: 1.1, rotate: 5 }}
                className="flex flex-col items-center space-y-2 cursor-pointer"
              >
                <div className={`p-4 rounded-full bg-white/10 backdrop-blur-sm border border-white/20 ${item.color}`}>
                  <item.icon size={32} />
                </div>
                <span className="text-sm text-gray-300 font-medium">{item.label}</span>
              </motion.div>
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default TechStackSection;