INSERT INTO sn.sn_batch_info (batch_name,batch_code,batch_number,batch_extra,work_code,product_code,UDI,sn_max,sn_min,status,comment,create_by,update_by,created_at,updated_at,deleted_at,product_month,image_file,`external`,product_id,sn_code_rules,batch_img_file,sn_format,sn_format_info,batch_code_format,batch_code_format_info) VALUES
	 ('','202207001',11,1,'work001','CWY-001','UDI001','8929CWY-001000012','8929CWY-001000001',1,'1',0,0,'2022-07-27 11:54:57.225000000','2022-07-27 11:54:57.225000000',NULL,'2022-07-03',NULL,0,1,0,'http://127.0.0.1:8000/static/uploadfile/6c41d246-b7f6-4e43-bce0-5eb5ac6f7f9c.jpg',0,'',0,NULL),
	 ('','202207002',11,2,'work002','CWY-001','UDI001','8929CWY-001000025','8929CWY-001000013',1,'2',0,0,'2022-07-27 11:55:22.015000000','2022-07-27 11:55:22.015000000',NULL,'2022-07-03',NULL,0,1,0,'http://127.0.0.1:8000/static/uploadfile/d20ed178-1848-41a1-80e3-751d78aa7b4a.jpg',0,'',0,NULL),
	 ('','LOT001',11,2,'work003','CWY-001','UDI001','8929CWY-001000038','8929CWY-001000026',1,'2',0,0,'2022-07-27 11:56:28.389000000','2022-07-27 12:01:33.340000000',NULL,'2022-07-03',NULL,0,1,0,'http://127.0.0.1:8000/static/uploadfile/c2b8b5e9-9eb5-4c8f-9d21-4ab332248405.jpg',0,'',1,'LOT001'),
	 ('','202207003',11,11,'11','CWY-001','UDI001','8929CWY-001000060','8929CWY-001000039',1,'11',0,0,'2022-07-27 13:33:42.080000000','2022-07-27 13:33:42.080000000',NULL,'2022-07-03',NULL,0,1,0,'http://127.0.0.1:8000/static/uploadfile/9070349a-52e4-4a8b-96d8-f5436b655c40.jpg',0,'',0,''),
	 ('','202207007',11,11,'1','CWY-001','UDI001','8929CWY-001000115','8929CWY-001000094',3,'2',0,0,'2022-07-27 13:34:34.698000000','2022-07-27 13:39:30.183000000',NULL,'2022-07-03',NULL,1,1,0,'http://127.0.0.1:8000/static/uploadfile/597a867c-03db-4377-abf4-f15049c830d7.jpg',0,'(02)',0,''),
	 ('','202207005',11,1,'work001','CWY-001','UDI001','LOT023','LOT001',1,'2',0,0,'2022-07-27 13:37:38.632000000','2022-07-27 13:37:38.632000000',NULL,'2022-07-03',NULL,1,1,1,'http://127.0.0.1:8000/static/uploadfile/2bceb4b0-8624-42fe-8bec-f17b2c2e3607.jpg',0,'',0,''),
	 ('','21202207007',12,12,'12','CWY-001','UDI001','218929CWY-001000117','218929CWY-001000094',1,'2',0,0,'2022-07-27 13:38:07.282000000','2022-07-27 13:38:50.389000000',NULL,'2022-07-03',NULL,1,1,0,'http://127.0.0.1:8000/static/uploadfile/82b54095-2d12-4ae6-a0d4-20d9aafee968.jpg',1,'21',0,'');


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
  `sn_format_info` varchar(255) DEFAULT NULL,
  `batch_code_format` int DEFAULT NULL,
  `batch_code_format_info` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`batch_id`),
  KEY `idx_batch_info_create_by` (`create_by`),
  KEY `idx_batch_info_update_by` (`update_by`),
  KEY `idx_batch_info_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=112 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


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
  `sn_format_info` varchar(255) DEFAULT NULL,
  `batch_code_format` int DEFAULT NULL,
  `batch_code_format_info` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`batch_id`),
  KEY `idx_batch_info_create_by` (`create_by`),
  KEY `idx_batch_info_update_by` (`update_by`),
  KEY `idx_batch_info_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=112 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


-- sn.sn_product_info definition

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
) ENGINE=InnoDB AUTO_INCREMENT=38 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO sn.sn_product_info (product_name,product_code,UDI,status,comment,create_by,update_by,created_at,updated_at,deleted_at,image_file) VALUES
	 ('测温仪','CWY-001','UDI001',0,'测温仪产品',0,NULL,'2022-07-05 22:01:33.038000000','2022-07-05 22:01:33.038000000',NULL,'http://127.0.0.1:8000/static/uploadfile/960a304e251f95cac9a8313fc3177f3e660952dd.jfif'),
	 ('血压仪','XYY-001','UDI002',0,'血压仪产品',0,NULL,'2022-07-05 22:01:33.038000000','2022-07-05 22:01:33.038000000',NULL,'http://127.0.0.1:8000/static/uploadfile/5d718947-bc00-4c4b-9bad-975d21c37058.jpg
');

INSERT INTO sn.sys_dict_data
(dict_code, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, `default`, remark, create_by, update_by, created_at, updated_at, deleted_at)
VALUES(34, 1, '新建', '5', 'sn_batch_status', '', '', '', 2, '', '新建批次', 1, 1, '2021-05-13 19:56:40.845000000', '2021-05-13 19:56:40.845000000', NULL);
INSERT INTO sn.sys_dict_data
(dict_code, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, `default`, remark, create_by, update_by, created_at, updated_at, deleted_at)
VALUES(35, 1, '生产中', '1', 'sn_batch_status', '', '', '', 2, '', '生产中批次', 1, 1, '2021-05-13 19:56:40.845000000', '2021-05-13 19:56:40.845000000', NULL);
INSERT INTO sn.sys_dict_data
(dict_code, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, `default`, remark, create_by, update_by, created_at, updated_at, deleted_at)
VALUES(36, 1, '完成', '2', 'sn_batch_status', '', '', '', 2, '', '生产完成', 1, 1, '2021-05-13 19:56:40.845000000', '2021-05-13 19:56:40.845000000', NULL);
INSERT INTO sn.sys_dict_data
(dict_code, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, `default`, remark, create_by, update_by, created_at, updated_at, deleted_at)
VALUES(37, 1, '已入库', '3', 'sn_batch_status', '', '', '', 2, '', '已经入库', 1, 1, '2021-05-13 19:56:40.845000000', '2021-05-13 19:56:40.845000000', NULL);
INSERT INTO sn.sys_dict_data
(dict_code, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, `default`, remark, create_by, update_by, created_at, updated_at, deleted_at)
VALUES(38, 1, '取消', '4', 'sn_batch_status', '', '', '', 2, '', '取消批次', 1, 1, '2021-05-13 19:56:40.845000000', '2021-05-13 19:56:40.845000000', NULL);
INSERT INTO sn.sys_dict_data
(dict_code, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, `default`, remark, create_by, update_by, created_at, updated_at, deleted_at)
VALUES(39, 1, '自制', '0', 'sn_batch_external', '', '', '', 2, '', '自制', 1, 1, '2021-05-13 19:56:40.845000000', '2021-05-13 19:56:40.845000000', NULL);
INSERT INTO sn.sys_dict_data
(dict_code, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, `default`, remark, create_by, update_by, created_at, updated_at, deleted_at)
VALUES(40, 1, '外购', '1', 'sn_batch_external', '', '', '', 2, '', '外购', 1, 1, '2021-05-13 19:56:40.845000000', '2021-05-13 19:56:40.845000000', NULL);


/* SN详细信息 */
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

DELETE FROM sys_dict_data where dict_type='sn_batch_status';
DELETE FROM sys_dict_data where dict_type='sn_info_status';

