<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">
        <el-form ref="queryForm" :model="queryParams" :inline="true" label-width="120px">
          <el-form-item label="批号(LOT)" prop="batchCode">
            <el-input
              v-model="queryParams.batchCode"
              placeholder="请输入批号"
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>
          <el-form-item label="产品型号" prop="productCode">
            <el-input
              v-model="queryParams.productCode"
              placeholder="请输入产品型号"
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">搜索</el-button>
            <el-button icon="el-icon-refresh" size="mini" @click="resetQuery">重置</el-button>
          </el-form-item>
        </el-form>

        <el-row :gutter="10" class="mb8">
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:export']"
              type="warning"
              icon="el-icon-plus"
              size="mini"
              @click="doBox"
            >开始装箱</el-button>
          </el-col>
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:export']"
              type="primary"
              icon="el-icon-setting"
              size="mini"
              @click="exWarehouse"
            >出库</el-button>
          </el-col>
        </el-row>

        <el-table v-loading="loading" :data="boxList" border >
          <el-table-column label="箱号" width="60" align="center" prop="BoxId" />
          <el-table-column label="批号(LOT)" width="140" align="center" prop="BatchCode" />
          <el-table-column label="产品型号" width="80" align="center" prop="ProductCode" />
          <el-table-column label="UDI号" align="center" prop="UDI" />
          <el-table-column label="工单号" width="150" align="center" prop="WorkCode" />
          <el-table-column label="装箱数量" width="150" align="center" prop="BoxSum" />
          <el-table-column label="状态" width="80" align="center" prop="Status" :formatter="statusFormat">
            <template slot-scope="scope">
              <el-tag
                :type="scope.row.Status === 0 ? 'danger' : 'success'"
                disable-transitions
              >{{ statusFormat(scope.row) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="装箱时间" align="center" prop="createdAt" width="155">
            <template slot-scope="scope">
              <span>{{ parseTime(scope.row.createdAt) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" align="center" class-name="small-padding fixed-width">
          <template slot-scope="scope">
            <el-button
                v-permisaction="['admin:sysPost:edit']"
                size="mini"
                type="text"
                icon="el-icon-edit"
                @click="handleUpdateBox(scope.row)"
              >修改</el-button>
          </template>
        </el-table-column>
        </el-table>

        <pagination
          v-show="total>0"
          :total="total"
          :page.sync="queryParams.pageIndex"
          :limit.sync="queryParams.pageSize"
          @pagination="getList"
        />
        <!-- 添加或修改批次对话框 -->
        <el-dialog :title="title" :visible.sync="open" width="600px">
          <el-form ref="form" :model="form" :rules="rules" label-width="120px">
            <el-form-item label="装箱数量" prop="BoxSum">
                <el-input v-model="form.BoxSum" placeholder="装箱数量" />
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button type="primary" @click="submitForm">确 定</el-button>
            <el-button @click="cancel">取 消</el-button>
          </div>
        </el-dialog>
      </el-card>
    </template>
  </BasicLayout>
</template>

<script>
import { listPost, updatePost } from '@/api/sn/box-info'
import { formatJson } from '@/utils'
import moment from 'moment'

export default {
  name: 'BoxInfoManage',
  data() {
    return {
      // 遮罩层
      loading: true,
      // 选中数组
      ids: [],
      // 非单个禁用
      single: true,
      // 非多个禁用
      multiple: true,
      // 总条数
      total: 0,
      // 岗位表格数据
      batchList: [],
      // 弹出层标题
      title: '',
      // 是否显示弹出层
      open: false,
      // 状态数据字典
      statusOptions: [],
      // 查询参数
      queryParams: {
        pageIndex: 1,
        pageSize: 10,
        postCode: undefined,
        postName: undefined,
        status: undefined
      },
      // 表单参数
      form: {
      },
      // 表单校验
      rules: {

      }
    }
  },
  created() {
    this.getList()
    this.getDicts('box_info_status').then(response => {
      this.statusOptions = response.data
    })
  },
  methods: {

    /** 查询装箱列表 */
    getList() {
      this.loading = true
      listPost(this.queryParams).then(response => {
        this.boxList = response.data.list
        this.total = response.data.count
        this.loading = false
      })
    },

    // 箱子状态翻译
    statusFormat(row) {
      return this.selectDictLabel(this.statusOptions, row.Status)
    }, 
    
    // 取消按钮
    cancel() {
      this.open = false
      this.reset()
    },
    // 表单重置
    reset() {
      this.form = {
        BoxId: undefined
      }
      this.resetForm('form')
    },
    /** 搜索按钮操作 */
    handleQuery() {
      this.queryParams.pageIndex = 1
      this.getList()
    },
    /** 重置按钮操作 */
    resetQuery() {
      this.resetForm('queryForm')
      this.handleQuery()
    },
    /** 修改按钮操作 */
    handleUpdateBox(row) {
        this.reset()
        this.form.BoxId = row.BoxId
        this.open = true
        this.title = '修改装箱数量'
    },

    /** 提交按钮 */
    submitForm: function() {
        this.form.BoxSum = parseInt(this.form.BoxSum)
        updatePost(this.form, this.form.BoxId).then(response => {
            if (response.code === 200) {
                this.msgSuccess(response.msg)
                this.open = false
                this.getList()
            } else {
                this.msgError(response.msg)
            }
        })

    },

    //执行装箱操作
    doBox() {
      this.$router.push({path: '/sn/box-relation-list'});
    },
    
    //执行出库操作
    exWarehouse() {
      this.$router.push({path: '/sn/ex-warehouse-manage'});
    }
  }
}
</script>
