-- NameカラムをExerciseTypeカラムに変更するマイグレーション
-- 既存データを保持したまま、カラムを変更する

-- 1. 新しいexercise_typeカラムを追加（INT型、デフォルト0=未指定）
ALTER TABLE workouts ADD COLUMN exercise_type INT NOT NULL DEFAULT 0;

-- 2. 既存のnameデータに基づいてexercise_typeを設定
UPDATE workouts SET exercise_type = 1 WHERE name = 'ベンチプレス' OR name LIKE '%ベンチプレス%';
UPDATE workouts SET exercise_type = 2 WHERE name = 'スクワット';
UPDATE workouts SET exercise_type = 3 WHERE name = 'デッドリフト';
UPDATE workouts SET exercise_type = 4 WHERE name = 'ダンベルショルダープレス' OR name LIKE '%ショルダープレス%';
UPDATE workouts SET exercise_type = 5 WHERE name = '懸垂';
UPDATE workouts SET exercise_type = 6 WHERE name = 'サイドレイズ';
UPDATE workouts SET exercise_type = 7 WHERE name = 'ワンハンドロー' OR name LIKE '%ロー%';
UPDATE workouts SET exercise_type = 8 WHERE name = 'ハイプル';

-- 3. nameカラムを削除
ALTER TABLE workouts DROP COLUMN name;

-- 4. exercise_typeカラムにインデックスを追加（検索性能向上）
CREATE INDEX idx_exercise_type ON workouts(exercise_type);

