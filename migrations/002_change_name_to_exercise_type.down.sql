-- exercise_typeをnameに戻すロールバック用マイグレーション

-- 1. インデックスを削除
DROP INDEX IF EXISTS idx_exercise_type;

-- 2. nameカラムを再追加
ALTER TABLE workouts ADD COLUMN name VARCHAR(255) NOT NULL DEFAULT '';

-- 3. exercise_typeから元のnameに戻す
UPDATE workouts SET name = 'ベンチプレス' WHERE exercise_type = 1;
UPDATE workouts SET name = 'スクワット' WHERE exercise_type = 2;
UPDATE workouts SET name = 'デッドリフト' WHERE exercise_type = 3;
UPDATE workouts SET name = 'ダンベルショルダープレス' WHERE exercise_type = 4;
UPDATE workouts SET name = '懸垂' WHERE exercise_type = 5;
UPDATE workouts SET name = 'サイドレイズ' WHERE exercise_type = 6;
UPDATE workouts SET name = 'ワンハンドロー' WHERE exercise_type = 7;
UPDATE workouts SET name = 'ハイプル' WHERE exercise_type = 8;
UPDATE workouts SET name = '未指定' WHERE exercise_type = 0;

-- 4. exercise_typeカラムを削除
ALTER TABLE workouts DROP COLUMN exercise_type;

