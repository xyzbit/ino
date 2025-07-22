import React from 'react';
import { motion } from 'framer-motion';
import { Search, MessageSquare, BarChart3, Shield, Zap, Globe } from 'lucide-react';

const FeaturesSection: React.FC = () => {
  const features = [
    {
      icon: Search,
      title: '智能检索',
      description: '基于向量数据库的语义检索，快速定位相关知识',
      gradient: 'from-blue-500 to-cyan-500',
      delay: 0.1,
    },
    {
      icon: MessageSquare,
      title: '对话生成',
      description: '结合上下文的智能对话，提供准确的问答服务',
      gradient: 'from-purple-500 to-pink-500',
      delay: 0.2,
    },
    {
      icon: BarChart3,
      title: '数据分析',
      description: '实时数据处理与分析，洞察业务趋势',
      gradient: 'from-green-500 to-emerald-500',
      delay: 0.3,
    },
    {
      icon: Shield,
      title: '安全可靠',
      description: '企业级安全保障，数据隐私全面保护',
      gradient: 'from-orange-500 to-red-500',
      delay: 0.4,
    },
    {
      icon: Zap,
      title: '高性能',
      description: '毫秒级响应速度，支持高并发访问',
      gradient: 'from-yellow-500 to-orange-500',
      delay: 0.5,
    },
    {
      icon: Globe,
      title: '多语言',
      description: '支持多种语言模型，满足全球化需求',
      gradient: 'from-indigo-500 to-purple-500',
      delay: 0.6,
    },
  ];

  return (
    <section className="py-20 bg-slate-900 relative overflow-hidden">
      {/* 背景装饰 */}
      <div className="absolute inset-0">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl" />
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
              核心功能
            </span>
          </h2>
          <p className="text-xl text-gray-300 max-w-3xl mx-auto">
            集成最新AI技术，为您提供全方位的智能服务体验
          </p>
        </motion.div>

        {/* 功能卡片网格 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <motion.div
              key={index}
              initial={{ opacity: 0, y: 50 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: feature.delay }}
              viewport={{ once: true }}
              whileHover={{ y: -10, scale: 1.02 }}
              className="group relative"
            >
              {/* 卡片背景 */}
              <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-white/5 rounded-2xl backdrop-blur-sm border border-white/10 group-hover:border-white/20 transition-all duration-300" />
              
              {/* 发光效果 */}
              <div className={`absolute inset-0 bg-gradient-to-br ${feature.gradient} opacity-0 group-hover:opacity-20 rounded-2xl blur-xl transition-all duration-500`} />
              
              <div className="relative p-8 h-full">
                {/* 图标 */}
                <motion.div
                  whileHover={{ rotate: 360 }}
                  transition={{ duration: 0.6 }}
                  className={`inline-flex p-4 rounded-2xl bg-gradient-to-br ${feature.gradient} mb-6 shadow-lg`}
                >
                  <feature.icon size={32} className="text-white" />
                </motion.div>
                
                {/* 内容 */}
                <h3 className="text-2xl font-bold text-white mb-4 group-hover:text-cyan-400 transition-colors">
                  {feature.title}
                </h3>
                
                <p className="text-gray-300 leading-relaxed">
                  {feature.description}
                </p>
                
                {/* 装饰线条 */}
                <motion.div
                  initial={{ width: 0 }}
                  whileInView={{ width: '100%' }}
                  transition={{ duration: 0.8, delay: feature.delay + 0.3 }}
                  viewport={{ once: true }}
                  className={`absolute bottom-0 left-0 h-1 bg-gradient-to-r ${feature.gradient} rounded-full`}
                />
              </div>
            </motion.div>
          ))}
        </div>

        {/* 底部CTA */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.8 }}
          viewport={{ once: true }}
          className="text-center mt-16"
        >
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="px-8 py-4 bg-gradient-to-r from-cyan-500 to-blue-600 text-white font-semibold rounded-full shadow-lg hover:shadow-cyan-500/25 transition-all duration-300"
          >
            探索更多功能
          </motion.button>
        </motion.div>
      </div>
    </section>
  );
};

export default FeaturesSection;