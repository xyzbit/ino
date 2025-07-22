import React from 'react';
import { motion } from 'framer-motion';
import { Github, Twitter, Linkedin, Mail, ExternalLink, Heart } from 'lucide-react';

const Footer: React.FC = () => {
  const footerSections = [
    {
      title: '产品',
      links: [
        { name: '功能特性', href: '#features' },
        { name: 'API 文档', href: '#docs' },
        { name: '定价方案', href: '#pricing' },
        { name: '更新日志', href: '#changelog' },
      ],
    },
    {
      title: '开发者',
      links: [
        { name: '快速开始', href: '#quickstart' },
        { name: 'SDK 下载', href: '#sdk' },
        { name: '示例代码', href: '#examples' },
        { name: '社区论坛', href: '#community' },
      ],
    },
    {
      title: '公司',
      links: [
        { name: '关于我们', href: '#about' },
        { name: '团队介绍', href: '#team' },
        { name: '联系我们', href: '#contact' },
        { name: '招聘信息', href: '#careers' },
      ],
    },
    {
      title: '资源',
      links: [
        { name: '技术博客', href: '#blog' },
        { name: '白皮书', href: '#whitepaper' },
        { name: '案例研究', href: '#cases' },
        { name: '帮助中心', href: '#help' },
      ],
    },
  ];

  const socialLinks = [
    { icon: Github, href: 'https://github.com', label: 'GitHub' },
    { icon: Twitter, href: 'https://twitter.com', label: 'Twitter' },
    { icon: Linkedin, href: 'https://linkedin.com', label: 'LinkedIn' },
    { icon: Mail, href: 'mailto:contact@kag-system.com', label: 'Email' },
  ];

  return (
    <footer className="bg-slate-900 border-t border-white/10">
      <div className="max-w-7xl mx-auto px-6">
        {/* 主要内容 */}
        <div className="py-16">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 gap-8">
            {/* 品牌信息 */}
            <div className="lg:col-span-2">
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6 }}
                viewport={{ once: true }}
              >
                {/* Logo */}
                <div className="flex items-center space-x-2 mb-6">
                  <div className="w-10 h-10 bg-gradient-to-br from-cyan-400 to-blue-500 rounded-xl flex items-center justify-center">
                    <span className="text-white font-bold text-lg">K</span>
                  </div>
                  <span className="text-2xl font-bold text-white">KAG System</span>
                </div>
                
                <p className="text-gray-300 mb-6 leading-relaxed">
                  智能知识增强生成系统，融合先进的AI技术与知识图谱，
                  为企业提供精准、高效的智能问答与内容生成服务。
                </p>
                
                {/* 社交链接 */}
                <div className="flex space-x-4">
                  {socialLinks.map((social, index) => (
                    <motion.a
                      key={index}
                      href={social.href}
                      target="_blank"
                      rel="noopener noreferrer"
                      whileHover={{ scale: 1.1, y: -2 }}
                      className="p-3 bg-white/10 backdrop-blur-sm rounded-lg border border-white/20 text-gray-300 hover:text-white hover:border-cyan-400/50 transition-all duration-300"
                      aria-label={social.label}
                    >
                      <social.icon size={20} />
                    </motion.a>
                  ))}
                </div>
              </motion.div>
            </div>
            
            {/* 链接分组 */}
            {footerSections.map((section, sectionIndex) => (
              <div key={sectionIndex}>
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.6, delay: sectionIndex * 0.1 }}
                  viewport={{ once: true }}
                >
                  <h3 className="text-white font-semibold mb-4">{section.title}</h3>
                  <ul className="space-y-3">
                    {section.links.map((link, linkIndex) => (
                      <li key={linkIndex}>
                        <motion.a
                          href={link.href}
                          whileHover={{ x: 5 }}
                          className="text-gray-300 hover:text-cyan-400 transition-colors duration-300 flex items-center group"
                        >
                          {link.name}
                          <ExternalLink size={14} className="ml-1 opacity-0 group-hover:opacity-100 transition-opacity" />
                        </motion.a>
                      </li>
                    ))}
                  </ul>
                </motion.div>
              </div>
            ))}
          </div>
        </div>
        
        {/* 分割线 */}
        <div className="border-t border-white/10" />
        
        {/* 底部信息 */}
        <div className="py-8">
          <motion.div
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            transition={{ duration: 0.6 }}
            viewport={{ once: true }}
            className="flex flex-col md:flex-row justify-between items-center space-y-4 md:space-y-0"
          >
            <div className="flex flex-col md:flex-row items-center space-y-2 md:space-y-0 md:space-x-6 text-gray-400 text-sm">
              <span>© 2024 KAG System. All rights reserved.</span>
              <div className="flex space-x-6">
                <a href="#privacy" className="hover:text-cyan-400 transition-colors">
                  隐私政策
                </a>
                <a href="#terms" className="hover:text-cyan-400 transition-colors">
                  服务条款
                </a>
                <a href="#cookies" className="hover:text-cyan-400 transition-colors">
                  Cookie 政策
                </a>
              </div>
            </div>
            
            <div className="flex items-center space-x-2 text-gray-400 text-sm">
              <span>Made with</span>
              <motion.div
                animate={{ scale: [1, 1.2, 1] }}
                transition={{ duration: 1, repeat: Infinity }}
              >
                <Heart size={16} className="text-red-500" fill="currentColor" />
              </motion.div>
              <span>by KAG Team</span>
            </div>
          </motion.div>
        </div>
      </div>
      
      {/* 背景装饰 */}
      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-cyan-500/50 to-transparent" />
    </footer>
  );
};

export default Footer;