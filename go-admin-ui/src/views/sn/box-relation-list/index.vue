<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">
        <el-form ref="queryForm" :model="queryParams" :inline="true" label-width="120px">
          <el-form-item label="箱号" prop="BoxId">
            <el-input
              v-model="queryParams.BoxId"
              placeholder=""
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>
          <el-form-item label="装箱数量" prop="BoxSum">
            <el-input
              v-model="queryParams.BoxSum"
              placeholder=""
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">搜索</el-button>
          </el-form-item>
        </el-form>

        <el-row :gutter="10" class="mb8">
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:export']"
              type="warning"
              icon="el-icon-microphone"
              size="mini"
              @click="warningAudio"
            ></el-button>
          </el-col>
        </el-row>

        <el-table v-loading="loading" :data="boxList" border >
          <el-table-column label="序号" width="120" align="center" prop="BoxRelationId" />
          <el-table-column label="箱号" width="120" align="center" prop="BoxId" />
          <el-table-column label="SN号" align="center" prop="SNCode" />
          <el-table-column label="扫码枪IP" width="120" align="center" prop="ScanSource" />
          <el-table-column label="创建时间" align="center" prop="createdAt" width="155">
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
import { listPost } from '@/api/sn/box-relation-info'
import { formatJson } from '@/utils'
import moment from 'moment'

export default {
  name: 'BoxRelationInfoManage',
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
    /** 搜索按钮操作 */
    handleQuery() {
      this.queryParams.pageIndex = 1
      this.getList()
    },
    
    /** 当SN码有异常时，触发告警声音 */
    warningAudio() {
      this.audio = new Audio()
      this.audio.src  = "http://159.75.213.231:8000/static/audios/do_wrong.mp3"
      this.audio.play()
    }
  }
}
</script>
