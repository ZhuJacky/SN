<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">
        <el-table v-loading="loading" :data="boxList" border >
          <el-table-column label="箱号" width="60" align="center" prop="BoxId" />
          <el-table-column label="批号(LOT)" width="140" align="center" prop="BatchCode" />
          <el-table-column label="产品型号" width="80" align="center" prop="ProductCode" />
          <el-table-column label="UDI号" align="center" prop="UDI" />
          <el-table-column label="工单号" width="150" align="center" prop="WorkCode" />
          <el-table-column label="装箱数量" width="150" align="center" prop="BoxSum" />
          <el-table-column label="出库时间" align="center" prop="createdAt" width="155">
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
import { listPost, updatePost } from '@/api/sn/box-info'
import { formatJson } from '@/utils'
import moment from 'moment'

export default {
  name: 'ExWarehouseManage',
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
    }
  }
}
</script>
