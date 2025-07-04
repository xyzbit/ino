-- 创建KAG数据库
CREATE DATABASE IF NOT EXISTS ino CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE ino;

-- 用户表
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) UNIQUE NOT NULL,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    avatar_url VARCHAR(500),
    preferences JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_email (email)
);

-- 知识域表
CREATE TABLE domains (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    domain_name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    config JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_domain_name (domain_name)
);

-- 文档表
CREATE TABLE documents (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    document_id VARCHAR(64) UNIQUE NOT NULL,
    domain_id BIGINT NOT NULL,
    title VARCHAR(500) NOT NULL,
    content_type VARCHAR(50),
    file_path VARCHAR(1000),
    metadata JSON,
    tags JSON,
    status ENUM('processing', 'completed', 'failed') DEFAULT 'processing',
    chunks_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (domain_id) REFERENCES domains(id),
    INDEX idx_document_id (document_id),
    INDEX idx_domain_id (domain_id),
    INDEX idx_status (status)
);

-- 对话记录表
CREATE TABLE conversations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    conversation_id VARCHAR(64) UNIQUE NOT NULL,
    domain_id BIGINT NOT NULL,
    user_id BIGINT,
    content JSON NOT NULL,
    tags JSON,
    processed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (domain_id) REFERENCES domains(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_conversation_id (conversation_id),
    INDEX idx_domain_id (domain_id),
    INDEX idx_user_id (user_id)
);

-- 反馈表
CREATE TABLE feedback (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    query_id VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL,
    feedback_type ENUM('positive', 'negative', 'neutral') NOT NULL,
    rating INT CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    context JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_query_id (query_id),
    INDEX idx_user_id (user_id),
    INDEX idx_feedback_type (feedback_type)
);

-- 检索记录表
CREATE TABLE search_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    query_id VARCHAR(64) UNIQUE NOT NULL,
    user_id BIGINT,
    domain_id BIGINT,
    query_text TEXT NOT NULL,
    search_config JSON,
    results JSON,
    response_time_ms INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (domain_id) REFERENCES domains(id),
    INDEX idx_query_id (query_id),
    INDEX idx_user_id (user_id),
    INDEX idx_domain_id (domain_id)
);

-- 插入默认数据
INSERT INTO domains (domain_name, description) VALUES 
('general', '通用知识域'),
('code-review', '代码评审领域'),
('documentation', '文档管理领域');

INSERT INTO users (user_id, username, email) VALUES 
('admin', 'Administrator', 'admin@ino.com'),
('system', 'System User', 'system@ino.com'); 