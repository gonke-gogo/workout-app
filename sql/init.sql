-- 筋トレ記録アプリのデータベース初期化（SQLチューニング対応版）
CREATE DATABASE IF NOT EXISTS workoutdb;
USE workoutdb;

-- ユーザーの作成と権限付与
CREATE USER IF NOT EXISTS 'workoutuser'@'%' IDENTIFIED BY 'workoutpass';
GRANT ALL PRIVILEGES ON workoutdb.* TO 'workoutuser'@'%';
FLUSH PRIVILEGES;

-- ワークアウトテーブル（最新版: exercise_type対応）
CREATE TABLE IF NOT EXISTS workouts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    exercise_type INT NOT NULL DEFAULT 0 COMMENT '0:未指定, 1:ベンチプレス, 2:スクワット, 3:デッドリフト, 4:ショルダープレス, 5:懸垂, 6:サイドレイズ, 7:ワンハンドロー, 8:ハイプル',
    description TEXT,
    status TINYINT NOT NULL DEFAULT 0 COMMENT '0:予定, 1:実行中, 2:完了, 3:スキップ',
    difficulty TINYINT NOT NULL DEFAULT 1 COMMENT '1:初心者, 2:中級者, 3:上級者, 4:化け物',
    muscle_group BIGINT NOT NULL DEFAULT 0 COMMENT '0:未指定, 1:胸, 2:背中, 3:脚, 4:肩, 5:腕, 6:腹筋, 7:体幹, 8:臀部, 9:有酸素, 10:全身',
    sets INT UNSIGNED DEFAULT 3,
    reps INT UNSIGNED DEFAULT 10,
    weight DECIMAL(6,2) DEFAULT 0.00,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    
    -- データ整合性制約
    CHECK (status >= 0 AND status <= 3),
    CHECK (difficulty >= 1 AND difficulty <= 4),
    CHECK (sets > 0 AND sets <= 100),
    CHECK (reps > 0 AND reps <= 1000),
    CHECK (weight >= 0 AND weight <= 9999.99)
);

-- 効率的なインデックス戦略（パフォーマンス最適化版）

-- 1. exercise_typeインデックス（種目での検索用）
CREATE INDEX idx_exercise_type ON workouts(exercise_type);

-- 2. 複合インデックス（最も頻繁に使用される検索条件）
CREATE INDEX idx_workouts_status_difficulty ON workouts(status, difficulty);
CREATE INDEX idx_workouts_muscle_difficulty ON workouts(muscle_group, difficulty);
CREATE INDEX idx_workouts_status_muscle ON workouts(status, muscle_group);

-- 3. カバリングインデックス（SELECT句の最適化）
CREATE INDEX idx_workouts_covering ON workouts(status, muscle_group, difficulty, exercise_type, created_at);

-- 4. 完了日時インデックス（NULL値含む）
CREATE INDEX idx_workouts_completed ON workouts(completed_at);

-- 5. 範囲検索用インデックス
CREATE INDEX idx_workouts_created_range ON workouts(created_at, status);

-- 6. 統計クエリ用の複合インデックス
CREATE INDEX idx_workouts_stats ON workouts(status, muscle_group, weight, created_at);

-- 楽しいサンプルデータ（より多様なデータでチューニング効果を確認）
-- exercise_type: 1=ベンチプレス, 2=スクワット, 3=デッドリフト, 4=ショルダープレス, 5=懸垂, 6=サイドレイズ, 7=ワンハンドロー, 8=ハイプル
INSERT INTO workouts (exercise_type, description, status, difficulty, muscle_group, sets, reps, weight, notes) VALUES
-- 胸のワークアウト (muscle_group = 1)
(1, '胸をバキバキに鍛える！モテ男への第一歩', 0, 2, 1, 3, 10, 60.00, '今日は調子がいい！💪'),
(1, '上部胸筋を重点的に鍛える', 0, 3, 1, 4, 8, 45.00, '上部胸筋が効いてる！'),
(1, '下部胸筋を鍛えて厚みを出す', 0, 3, 1, 3, 12, 50.00, '下部胸筋がバキバキ！'),
(1, '胸の内側を集中的に鍛える', 0, 2, 1, 3, 15, 20.00, '胸の内側が効いてる！'),

-- 背中のワークアウト (muscle_group = 2)
(3, '背中を広くして逆三角形に！', 0, 3, 2, 3, 8, 80.00, '化け物級の重量に挑戦！🔥'),
(7, '広い背中で逆三角形を目指す', 0, 2, 2, 3, 12, 45.00, '背中が広がってきた！'),
(7, '背中の厚みを増す', 0, 2, 2, 4, 10, 40.00, '背中が厚くなってきた！'),
(7, '背中の中央部を鍛える', 0, 1, 2, 3, 15, 30.00, '背中の中央が効いてる！'),

-- 脚のワークアウト (muscle_group = 3)
(2, '下半身を鍛えてモテ男に！', 0, 1, 3, 4, 15, 0.00, '自重でもキツい...😅'),
(2, '太ももを太くして逞しく', 0, 2, 3, 4, 12, 100.00, '脚が太くなってきた！'),
(2, '太もも前部を集中的に鍛える', 0, 1, 3, 3, 20, 25.00, '太もも前部が効いてる！'),
(2, '太もも後部を鍛える', 0, 1, 3, 3, 15, 20.00, '太もも後部が効いてる！'),

-- 肩のワークアウト (muscle_group = 4)
(4, '肩を大きくして逞しく見せる', 0, 2, 4, 3, 10, 30.00, '肩が丸くなってきた💪'),
(6, '肩の外側を鍛える', 0, 1, 4, 3, 15, 8.00, '肩の外側が効いてる！'),
(4, '肩の後ろ側を鍛える', 0, 2, 4, 3, 12, 12.00, '肩の後ろ側が効いてる！'),
(4, '肩の前側を鍛える', 0, 1, 4, 3, 12, 10.00, '肩の前側が効いてる！'),

-- 腕のワークアウト (muscle_group = 5)
(1, '腕を太くして彼女にアピール', 0, 1, 5, 3, 20, 0.00, '筋肉痛で死にそう💀'),
(5, '腕を太くして逞しく', 0, 1, 5, 3, 15, 15.00, '腕が太くなってきた💪'),
(5, '腕の後ろ側を鍛える', 0, 2, 5, 3, 12, 35.00, '腕の後ろ側が効いてる！'),
(5, '腕の外側を鍛える', 0, 2, 5, 3, 12, 18.00, '腕の外側が効いてる！'),

-- 腹筋のワークアウト (muscle_group = 6)
(0, '腹筋を割ってビーチボディに！', 0, 1, 6, 3, 60, 0.00, '1分間が永遠に感じる⏰'),
(0, '腹筋を割ってモテ男に', 0, 1, 6, 3, 20, 0.00, '腹筋が少し見えてきた😊'),
(0, '下腹部を集中的に鍛える', 0, 2, 6, 3, 15, 0.00, '下腹部が効いてる！'),
(0, '腹斜筋を鍛える', 0, 2, 6, 3, 45, 0.00, '腹斜筋が効いてる！'),

-- 完了済みのワークアウト（統計クエリ用）
(1, '胸をバキバキに鍛えた！', 2, 2, 1, 3, 10, 65.00, '記録更新！💪'),
(2, '下半身を鍛えた！', 2, 1, 3, 4, 15, 0.00, '筋肉痛で歩けない😅'),
(3, '背中を鍛えた！', 2, 3, 2, 3, 8, 85.00, '化け物級の重量達成！🔥'),

-- スキップしたワークアウト
(0, '有酸素運動で脂肪燃焼', 3, 1, 9, 1, 30, 0.00, '雨でスキップ😢'),
(0, '柔軟性を高める', 3, 1, 7, 1, 60, 0.00, '時間がなくてスキップ😅');

-- 統計情報の表示
SELECT 
    '筋トレ記録アプリ（SQLチューニング版）' as app_name,
    COUNT(*) as total_workouts,
    COUNT(CASE WHEN status = 2 THEN 1 END) as completed_workouts,
    COUNT(CASE WHEN status = 3 THEN 1 END) as skipped_workouts,
    '楽しく筋トレしましょう！💪' as message
FROM workouts;

-- インデックス使用状況の確認用クエリ例
SELECT 'インデックス分析用クエリ例:' as info;
SELECT '1. 複合インデックス効果確認:' as query_type;
SELECT '   EXPLAIN SELECT * FROM workouts WHERE status = 0 AND difficulty = 2;' as query;
SELECT '2. カバリングインデックス効果確認:' as query_type;
SELECT '   EXPLAIN SELECT status, muscle_group, difficulty, exercise_type FROM workouts WHERE status = 0;' as query;
SELECT '3. 範囲検索効果確認:' as query_type;
SELECT '   EXPLAIN SELECT * FROM workouts WHERE created_at >= "2024-01-01" AND status = 0;' as query;
