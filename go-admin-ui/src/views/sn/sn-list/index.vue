<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">
        <el-form ref="queryForm" :model="queryParams" :inline="true" label-width="68px">
          <el-form-item label="批号" prop="postCode">
            <el-input
              v-model="queryParams.postCode"
              placeholder="请输入批次编码"
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>
          <el-form-item label="批次名称" prop="postName">
            <el-input
              v-model="queryParams.postName"
              placeholder="请输入批次名称"
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>
          <el-form-item label="状态" prop="status">
            <el-select v-model="queryParams.status" placeholder="批次状态" clearable size="small">
              <el-option
                v-for="dict in statusOptions"
                :key="dict.value"
                :label="dict.label"
                :value="dict.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">搜索</el-button>
            <el-button icon="el-icon-refresh" size="mini" @click="resetQuery">重置</el-button>
          </el-form-item>
        </el-form>

        <el-row :gutter="10" class="mb8">
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:add']"
              type="primary"
              icon="el-icon-plus"
              size="mini"
              @click="handleAdd"
            >新增</el-button>
          </el-col>
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:edit']"
              type="success"
              icon="el-icon-edit"
              size="mini"
              :disabled="single"
              @click="handleUpdate"
            >修改</el-button>
          </el-col>
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:remove']"
              type="danger"
              icon="el-icon-delete"
              size="mini"
              :disabled="multiple"
              @click="handleDelete"
            >删除</el-button>
          </el-col>
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:export']"
              type="warning"
              icon="el-icon-download"
              size="mini"
              @click="handleExport"
            >导出</el-button>
          </el-col>
          <el-col :span="1.5">
              <el-button
                v-permisaction="['admin:sysPost:export']"
                type="warning"
                icon="el-icon-view"
                size="mini"
                @click="handleDetails"
              >详情</el-button>
            </el-col>
        </el-row>

        <el-table v-loading="loading" :data="batchList" border @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="55" align="center" />
          <el-table-column label="批号" width="100" align="center" prop="BatchCode" />
          <el-table-column label="批次名称" width="180" align="center" prop="BatchName" />
          <el-table-column label="产品型号" width="100" align="center" prop="ProductCode" />
          <el-table-column label="UDI号" width="150" align="center" prop="UDI" />
          <el-table-column label="工单号" width="150" align="center" prop="WorkCode" />
          <el-table-column label="SN编码" width="180" align="center" prop="SNCode" />
          <el-table-column label="生产月份" align="center" prop="ProductMonth" :formatter="dateFormat">
            <template slot-scope="scope">
              <el-tag>{{ dateFormat(scope.row) }}</el-tag>
            </template>
          </el-table-column>

          <el-table-column label="状态" align="center" prop="status" :formatter="statusFormat">
            <template slot-scope="scope">
              <el-tag
                :type="scope.row.status === 1 ? 'danger' : 'success'"
                disable-transitions
              >{{ statusFormat(scope.row) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" align="center" prop="createdAt" width="180">
            <template slot-scope="scope">
              <span>{{ parseTime(scope.row.createdAt) }}</span>
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
      </el-card>
    </template>
  </BasicLayout>
</template>

<script>
import { listPost, getPost, delPost, addPost, updatePost } from '@/api/sn/sn-info'
import { formatJson } from '@/utils'
import moment from 'moment'

export default {
  name: 'SysPostManage',
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
      // 外协
      externalOptions: [],
      // 查询参数
      queryParams: {
        pageIndex: 1,
        pageSize: 10,
        postCode: undefined,
        postName: undefined,
        status: undefined
      },
      // 表单参数
      form: {},
      // 表单校验
      rules: {
        BatchName: [
          { required: true, message: '批次名称不能为空', trigger: 'blur' }
        ],
        ProductCode: [
          { required: true, message: '产品型号不能为空', trigger: 'blur' }
        ],
        BatchNumber: [
          { required: true, message: '批次数量不能为空', trigger: 'blur' }
        ]
      }
    }
  },
  created() {
    this.getList()
    this.getDicts('sys_normal_disable').then(response => {
      this.statusOptions = response.data
    })
  },
  methods: {
    /** 查询岗位列表 */
    getList() {
      this.loading = true
      listPost(this.queryParams).then(response => {
        this.batchList = response.data.list
        this.total = response.data.count
        this.loading = false
      })
    },
    // SN状态翻译
    statusFormat(row) {
      return this.selectDictLabel(this.statusOptions, row.status)
    },
    dateFormat(row) {
      return moment(row.ProductMonth).format("YYYY-MM")
    },
    // 取消按钮
    cancel() {
      this.open = false
      this.reset()
    },
    // 表单重置
    reset() {
      this.form = {
        BatchId: undefined,
        postCode: undefined,
        postName: undefined,
        sort: 0,
        status: '1',
        remark: undefined
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
      var query=this.$route.query;
      let batch_code = query.batch_code;
      this.queryParams.postCode = batch_code;
      //alert(batch_code)
    },
    // 多选框选中数据
    handleSelectionChange(selection) {
      this.ids = selection.map(item => item.BatchId)
      this.single = selection.length !== 1
      this.multiple = !selection.length
    },
    /** 新增按钮操作 */
    handleAdd() {
      this.reset()
      this.open = true
      this.title = '添加批次'
    },
    /** 修改按钮操作 */
    handleUpdate(row) {
      this.reset()

//      const postId = row.BatchId || this.ids
      /*const postId=row.BatchId || this.Ids*/
      const postId = (row.BatchId && [row.BatchId]) || this.ids
      getPost(postId).then(response => {
        this.form = response.data
        this.form.status = String(this.form.status)
        this.open = true
        this.title = '修改批次'
      })
    },
    handleDetails(row) {
      var query=this.$route.query;
      let batch_code = query.batch_code;
      },

    /** 提交按钮 */
    submitForm: function() {
      this.$refs['form'].validate(valid => {
        if (valid) {
          this.form.status = parseInt(this.form.status)
          if (this.form.BatchId !== undefined) {
            this.form.ProductMonth=this.form.ProductMonth.slice(0,7)
            updatePost(this.form, this.form.BatchId).then(response => {
              if (response.code === 200) {
                this.msgSuccess(response.msg)
                this.open = false
                this.getList()
              } else {
                this.msgError(response.msg)
              }
            })
          } else {
            addPost(this.form).then(response => {
              if (response.code === 200) {
                this.msgSuccess(response.msg)
                this.open = false
                this.getList()
              } else {
                this.msgError(response.msg)
              }
            })
          }
        }
      })
    },
    /** 删除按钮操作 */
    handleDelete(row) {
      // const postIds = row.postId || this.ids
      const Ids = (row.BatchId && [row.BatchId]) || this.ids
      this.$confirm('是否确认删除批次编号为"' + Ids + '"的数据项?', '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(function() {
        return delPost({ 'ids': Ids })
      }).then((response) => {
        if (response.code === 200) {
          this.msgSuccess(response.msg)
          this.open = false
          this.getList()
        } else {
          this.msgError(response.msg)
        }
      }).catch(function() {})
    },
    /** 导出按钮操作 */
    handleExport() {
      // const queryParams = this.queryParams
      this.$confirm('是否确认导出所有批次数据项?', '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.downloadLoading = true
        import('@/vendor/Export2Excel').then(excel => {
          const tHeader = ['批次ID', '批次编码', '批次名称', '创建时间']
          const filterVal = ['BatchId', 'BatchCode', 'BatchName', 'createdAt']
          const list = this.batchList
          const data = formatJson(filterVal, list)
          excel.export_json_to_excel({
            header: tHeader,
            data,
            filename: '批次管理',
            autoWidth: true, // Optional
            bookType: 'xlsx' // Optional
          })
          this.downloadLoading = false
        })
      }).catch(function() {})
    }
  }
}
</script>