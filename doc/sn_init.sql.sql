

DROP TABLE IF EXISTS sn_batch_info;

CREATE TABLE `sn_batch_info` (
  `batch_id` bigint NOT NULL AUTO_INCREMENT,
  `batch_name` varchar(128) DEFAULT NULL,
  `batch_code` varchar(128) DEFAULT NULL,
  `batch_number` int DEFAULT NULL,
  `batch_extra` int DEFAULT NULL,
  `work_code` varchar(128) DEFAULT NULL,
  `product_code` varchar(128) DEFAULT NULL,
  `UDI` varchar(128) DEFAULT NULL,
  `sn_max` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `sn_min` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  `comment` varchar(255) DEFAULT NULL,
  `create_by` bigint DEFAULT NULL COMMENT '创建者',
  `update_by` bigint DEFAULT NULL COMMENT '更新者',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最后更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '删除时间',
  `product_month` date DEFAULT NULL,
  `image_file` varchar(255) DEFAULT NULL,
  `external` tinyint(1) DEFAULT NULL,
  `product_id` bigint DEFAULT NULL,
  `sn_code_rules` int DEFAULT NULL,
  `batch_img_file` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `sn_format` int DEFAULT NULL,
  `udi_format_info` varchar(255) DEFAULT NULL,
  `lot_format_info` varchar(255) DEFAULT NULL,
  `sn_format_info` varchar(255) DEFAULT NULL,
  `batch_code_format` int DEFAULT NULL,
  `batch_code_format_info` varchar(100) DEFAULT NULL,
  `auto_sn_sum` int DEFAULT 0,
  PRIMARY KEY (`batch_id`),
  KEY `idx_batch_info_create_by` (`create_by`),
  KEY `idx_batch_info_update_by` (`update_by`),
  KEY `idx_batch_info_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


-- sn.sn_product_info definition

DROP TABLE IF EXISTS sn_product_info;

CREATE TABLE `sn_product_info` (
  `product_id` bigint NOT NULL AUTO_INCREMENT,
  `product_name` varchar(128) DEFAULT NULL,
  `product_code` varchar(128) DEFAULT NULL,
  `UDI` varchar(128) DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  `comment` varchar(255) DEFAULT NULL,
  `create_by` bigint DEFAULT NULL COMMENT '创建者',
  `update_by` bigint DEFAULT NULL COMMENT '更新者',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最后更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '删除时间',
  `image_file` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '产品图片',
  PRIMARY KEY (`product_id`),
  KEY `idx_batch_info_create_by` (`create_by`),
  KEY `idx_batch_info_update_by` (`update_by`),
  KEY `idx_batch_info_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


/* SN详细信息 */
DROP TABLE IF EXISTS sn_info;

CREATE TABLE `sn_info` (
  `sn_id` bigint NOT NULL AUTO_INCREMENT,
  `sn_code` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `batch_id` bigint DEFAULT NULL COMMENT 'batch_id',
  `batch_name` varchar(128) DEFAULT NULL,
  `batch_code` varchar(128) DEFAULT NULL,
  `work_code` varchar(128) DEFAULT NULL,
  `product_code` varchar(128) DEFAULT NULL,
  `UDI` varchar(128) DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  `create_by` bigint DEFAULT NULL COMMENT 'create_by',
  `update_by` bigint DEFAULT NULL COMMENT 'update_by',
  `created_at` datetime(3) DEFAULT NULL COMMENT 'created_at',
  `updated_at` datetime(3) DEFAULT NULL COMMENT 'updated_at',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT 'deleted_at',
  `product_month` date DEFAULT NULL,
  `product_id` bigint DEFAULT NULL,
  PRIMARY KEY (`sn_id`),
  KEY `idx_batch_info_create_by` (`create_by`),
  KEY `idx_batch_info_update_by` (`update_by`),
  KEY `idx_batch_info_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


DELETE FROM sys_dict_data where dict_type='sn_batch_status';
DELETE FROM sys_dict_data where dict_type='sn_info_status';
/* 批次状态 */
INSERT INTO sn.sys_dict_data (dict_sort,dict_label,dict_value,dict_type,css_class,list_class,is_default,status,`default`,remark,create_by,update_by,created_at,updated_at,deleted_at) VALUES
      (1,'新建','0','sn_batch_status','','','',2,'','新建批次',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'下单','1','sn_batch_status','','','',2,'','下单',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'检验','2','sn_batch_status','','','',2,'','检验',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'生产','3','sn_batch_status','','','',2,'','生产',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'完成','4','sn_batch_status','','','',2,'','完成',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL);

/* SN状态 */
INSERT INTO sn.sys_dict_data (dict_sort,dict_label,dict_value,dict_type,css_class,list_class,is_default,status,`default`,remark,create_by,update_by,created_at,updated_at,deleted_at) VALUES
      (1,'新建','0','sn_info_status','','','',2,'','新建SN',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'下单','1','sn_info_status','','','',2,'','下单',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'生产','2','sn_info_status','','','',2,'','生产',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'装箱','3','sn_info_status','','','',2,'','装箱',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'检验','4','sn_info_status','','','',2,'','检验',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'NG','5','sn_info_status','','','',2,'','NG',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'入库','6','sn_info_status','','','',2,'','入库',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
      (1,'出库','7','sn_info_status','','','',2,'','出库',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL);



/* 初始化产品列表 */

DELETE FROM sn_product_info;

INSERT INTO sn_product_info (product_code,product_name) VALUES ('500A','血氧仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('500B','血氧仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('500C','血氧仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('500D','血氧仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('500E','血氧仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('500F','血氧仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('500G','血氧仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('500H','血氧仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('900A','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('900W','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('HA100','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('HA101','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('HA120','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('HA121','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('HA200','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('HA200C','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('HA210','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('HA300','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('HW100A','血压计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('700A','体重称');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('BS200','体重称');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('BS201','体重称');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('BFS300','体脂秤');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('BFS200A','体脂称');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('BFS200B','体脂称');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('BFS200C','体脂称');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('BFS200D','体脂称');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('BFS711','体脂称');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR100+','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR202','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR203','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR205','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR300','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR301','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR302','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR400','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR402','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR403','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR409','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR409-BT','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR410','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR415','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR418','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('FR600','体温计');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100A','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100B','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100B+','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100B2','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100B3','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100B4','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100E','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100S','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100S+','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100S2','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100S4','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100S6','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100S8','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100S9','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('100T','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('200B2','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('200C','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('200S','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('MINI','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('SHA10','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('SHA20','胎心仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('300K','胎儿监护系统');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('700B','门诊产品');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('700C','门诊产品');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('ES100','理疗仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('ES200','理疗仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('ES210','理疗仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('ES220','理疗仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('800B','多参数监护仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('2000S','低频脉冲治疗仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('2000S1','低频脉冲治疗仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('300E','超声胎儿监护仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('300F','超声胎儿监护仪');
INSERT INTO sn_product_info (product_code,product_name) VALUES ('300P','超声胎儿监护仪');



DROP TABLE IF EXISTS sn_box_info;
CREATE TABLE `sn_box_info` (
  `box_id` bigint NOT NULL AUTO_INCREMENT,
  `batch_id` bigint DEFAULT NULL COMMENT 'batch_id',
  `batch_code` varchar(128) DEFAULT NULL,
  `work_code` varchar(128) DEFAULT NULL,
  `product_code` varchar(128) DEFAULT NULL,
  `UDI` varchar(128) DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  `scan_source` varchar(128) DEFAULT NULL,
  `box_sum` int DEFAULT 10,
  `create_by` bigint DEFAULT NULL COMMENT 'create_by',
  `update_by` bigint DEFAULT NULL COMMENT 'update_by',
  `created_at` datetime(3) DEFAULT NULL COMMENT 'created_at',
  `updated_at` datetime(3) DEFAULT NULL COMMENT 'updated_at',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`box_id`),
  KEY `idx_batch_info_create_by` (`create_by`),
  KEY `idx_batch_info_update_by` (`update_by`),
  KEY `idx_batch_info_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


DROP TABLE IF EXISTS sn_box_relation;
CREATE TABLE `sn_box_relation` (
  `box_relation_id` bigint NOT NULL AUTO_INCREMENT,
  `box_id` bigint DEFAULT NULL COMMENT 'box_id',
  `sn_code` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `scan_source` varchar(128) DEFAULT NULL,
  `batch_code` varchar(128) DEFAULT NULL,
  `product_code` varchar(128) DEFAULT NULL,
  `box_sum` int DEFAULT 10,
  `create_by` bigint DEFAULT NULL COMMENT 'create_by',
  `update_by` bigint DEFAULT NULL COMMENT 'update_by',
  `created_at` datetime(3) DEFAULT NULL COMMENT 'created_at',
  `updated_at` datetime(3) DEFAULT NULL COMMENT 'updated_at',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`box_relation_id`),
  KEY `sn_code` (`sn_code`),
  KEY `idx_batch_info_create_by` (`create_by`),
  KEY `idx_batch_info_update_by` (`update_by`),
  KEY `idx_batch_info_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


INSERT INTO sn.sys_dict_data (dict_sort,dict_label,dict_value,dict_type,css_class,list_class,is_default,status,`default`,remark,create_by,update_by,created_at,updated_at,deleted_at) VALUES
  (1,'装箱','0','box_info_status','','','',2,'','装箱',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
  (1,'入库','1','box_info_status','','','',2,'','入库',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
  (1,'出库','2','box_info_status','','','',2,'','出库',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL);

INSERT INTO sn.sys_dict_data (dict_sort,dict_label,dict_value,dict_type,css_class,list_class,is_default,status,`default`,remark,create_by,update_by,created_at,updated_at,deleted_at) VALUES
    (1,'自制','0','sn_batch_external','','','',0,'','自制',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL),
    (1,'外购','1','sn_batch_external','','','',0,'','外购',1,1,'2021-05-13 19:56:40.845000000','2021-05-13 19:56:40.845000000',NULL);