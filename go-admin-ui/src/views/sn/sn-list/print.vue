<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">
        <input  v-model="codeValue" placeholder="请输入条形码"/>
          <el-form-item>
            <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">搜索</el-button>
          </el-form-item>
        
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

        <el-table v-loading="loading" :data="SNList" border >
          <el-table-column type="selection" width="40" align="center" />
          <el-table-column label="SN号" align="center" prop="SNCode" />
          <el-table-column label="批次号" width="120" align="center" prop="BatchCode" />
          <el-table-column label="工单号" width="120" align="center" prop="WorkCode" />
          <el-table-column label="创建时间" align="center" prop="createdAt" width="155">
            <template slot-scope="scope">
              <span>{{ parseTime(scope.row.createdAt) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作">
        <template slot-scope="scope">
          <el-button type="primary" @click="doPrint(scope.row)">打印</el-button>
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
      <div ref="printDiv" style="display:none;" v-if="isReloadData">
                <div style="margin-top:7px;">
                    <img :src="SNCode | creatBarCode(printItem.SNCode)" style="width:auto; height:150px"/>
                </div>
                <div style="font-size:10px;margin-top:7px;">
                    产品序号：{{printItem.SNCode}}{{printItem.BatchCode}}
                </div>
                <div style="font-size:10px;margin-top:7px;">
                    批次编号：{{printItem.SNCode}}{{printItem.BatchCode}}
                </div>
                <div style="font-size:10px;margin-top:7px;">
                    工单编号：{{printItem.WorkCode}}
                </div>
       </div>
       <iframe ref="printIframe" frameborder="0" scrolling="no" style="margin: 0px;padding: 0px;width: 0px;height: 0px;"></iframe>

       </template>
  </BasicLayout>
</template>

<script>
//import { listPost,packBox } from '@/api/sn/box-relation-info'
import { listPost, getPost, updatePost } from '@/api/sn/sn-info'
import { formatJson } from '@/utils'
import moment from 'moment'
import getLodop from '@/utils/LodopFuncs'
export default {
  name: 'BoxRelationInfoManage',
  inject:['reload'], 
  data() {
    return {
              timer: '',
      goodCodeList:[
        　　　　 　　{uid: 1, brandName: '苹果',goodName: 'iphone13 Pro', code: '11112222333',price: 900, unit:'个', skuName: '8+526G', num: 1},
        ],
        isReloadData: true,
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
      SNList: [],
      // 是否显示弹出层
      open: false,
      // 状态数据字典
      statusOptions: [],
      // 查询参数
      queryParams: {
        pageIndex: 1,
        pageSize: 100,
        postCode: undefined,
        postName: undefined,
        status: undefined
      },
      boxData: {
        snCode: undefined,
        status: 3,
        scanSource: '666'
      },
      printItem:{},
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

    window.document.onkeypress = (e) => {
      if (window.event) { // IE
        this.nextCode = e.keyCode
      } else if (e.which) { // Netscape/Firefox/Opera
        this.nextCode = e.which
      }
 
      if (e.which === 13) { // 键盘回车事件
        if (this.code.length < 3) return // 扫码枪的速度很快，手动输入的时间不会让code的长度大于2，所以这里不会对扫码枪有效
        console.log('扫码结束。')
        console.log('条形码：', this.code)
        this.parseQRCode(this.code) // 获取到扫码枪输入的内容，做别的操作
        this.lastCode = ''
        this.lastTime = ''
        return
      }
 
      this.nextTime = new Date().getTime()
      if (!this.lastTime && !this.lastCode) {
        this.code = '' // 清空上次的条形码
        this.code += e.key
        console.log('扫码开始---', this.code)
      }
      if (this.lastCode && this.lastTime && this.nextTime - this.lastTime > 500) { // 当扫码前有keypress事件时,防止首字缺失
        this.code = e.key
        console.log('防止首字缺失。。。', this.code)
      } else if (this.lastCode && this.lastTime) {
        this.code += e.key
        console.log('扫码中。。。', this.code)
      }
      this.lastCode = this.nextCode
      this.lastTime = this.nextTime
    }
  },
  methods: {
     reload() {
      this.isReloadData = false
      this.$nextTick(() => {
        this.isReloadData = false
        this.isReloadData = true
      })
    },
    creatBarCode(barCodeData, printData) {
            console.log("触发条码生成事件");
            console.log(printData);
            let canvas = document.createElement("canvas");
            JsBarcode(canvas, barCodeData, {
                format: "CODE128",
                displayValue: true,
                margin: 0,
                height: 125,
                width: 2,
                fontSize: 30,
                textMargin: 10,
            });
            return canvas.toDataURL("image/png");
        },
       toClick() {        // 商品条码
                const LODOP = getLodop()
                if (LODOP) {
                    this.goodCodeList.forEach( (item,index) => {　　// 这里为打印的商品条码
                            // 反之，则以商品的数量作为打印份数依据。（注：这里是该商品条码打印的份数）
                            for( let i=0; i<item.num; i++) {
                                var strBodyStyle = '<style>'
                                strBodyStyle += '.dayinID { width: 200px;height: 155px;border: 1px solid #000;margin-top: 0px;}'
                                strBodyStyle += '.row5,.row1 { width: 100%;height: 20px;line-height: 20px;color: #000;text-align: center;font-weight: 700;font-size: 0.8rem;}'
                                // strBodyStyle += '.row1>span { width: 50%;height: 100%;text-align: center;line-height: 30px;color: #000;float: left;}'
                                strBodyStyle += '.row2 { width: 100%;height: 25px;text-align: center;color: #000;font-weight: 700;font-size: 1.2rem;line-height: 25px;}'
                                strBodyStyle += '.row3 { width: 100%;height: 20px;line-height: 20px;text-align: center;font-size: 0.8rem;}'
                                strBodyStyle += '.row4 { width: 100%;height: 60px;}'
                                strBodyStyle += '.tiaoma-space{width: 100%;height: 10px;margin-top: 20px;}'
                                strBodyStyle += '</style>' // 设置打印样式
                                var strFormHtml = strBodyStyle + '<body>' + document.getElementById('dayinID_'+index).innerHTML + '</body>'   // 获取打印内容
                           LODOP.SET_LICENSES("","EE0887D00FCC7D29375A695F728489A6","C94CEE276DB2187AE6B65D56B3FC2848","")
                                LODOP.PRINT_INIT('')  //初始化
                                LODOP.SET_PRINT_PAGESIZE(3, 790, 0, '')  // 设置横向(四个参数：打印方向及纸张类型（0(或其它)：打印方向由操作者自行选择或按打印机缺省设置；1：纵向打印,固定纸张；2：横向打印，固定纸张；3：纵向打印，宽度固定，高度按打印内容的高度自适应。），纸张宽(mm)，纸张高(mm),纸张名(必须纸张宽等于零时本参数才有效。))
                           
                                LODOP.ADD_PRINT_HTM('1%', '1%', '98%', '98%', strFormHtml)        // 设置打印内容
                                // LODOP.ADD_PRINT_TEXT('1%', '1%', '98%', '98%', strFormHtml)    // 设置打印内容
//                                LODOP.ADD_PRINT_BARCODE( 85, 55, 230, 60, '128Auto', item.code);   // 条码（六个参数：Top,Left,Width,Height,BarCodeType,BarCodeValue）
                                LODOP.ADD_PRINT_BARCODE( 85, 155, 230, 60, 'QRCode', item.code);   // 条码（六个参数：Top,Left,Width,Height,BarCodeType,BarCodeValue）
                                // LODOP.SET_PREVIEW_WINDOW(2, 0, 0, 800, 600, '')  // 设置预览窗口模式和大小
                                LODOP.PREVIEW()  // 预览。（这里可以进行预览，注意这里打开时记得把下面的print先注释。）另外，这里的预览只显示单个商品 打印出来的效果即该预览效果。
                                //LODOP.PRINT();　　// 打印a
                            }
                    });
                }
            },
doPrint(printItem) {
            var strFormHtml='<div data-v-54ecf1d8="" data-v-43eac8e8="" style=""><img data-v-54ecf1d8="" data-v-43eac8e8="" style="width: auto; height: 50px;"></div><div data-v-54ecf1d8="" data-v-43eac8e8="" style="font-size: 10px; margin-top: 10px;"> 产品序号：'+printItem.SNCode+'</div><div data-v-54ecf1d8="" data-v-43eac8e8="" style="font-size: 10px; margin-top: 3px;"> 批次编号：'+printItem.BatchCode+'</div><div data-v-54ecf1d8="" data-v-43eac8e8="" style="font-size: 10px; margin-top: 3px;"> 工单编号：'+printItem.WorkCode+'</div>'
            console.log(strFormHtml);
                const LODOP = getLodop()
                if (LODOP) {
                  LODOP.SET_LICENSES("","EE0887D00FCC7D29375A695F728489A6","C94CEE276DB2187AE6B65D56B3FC2848","")
                                LODOP.PRINT_INIT('')  //初始化
                                LODOP.SET_PRINT_PAGESIZE(3, 290, 20, 'abc')  // 设置横向(四个参数：打印方向及纸张类型（0(或其它)：打印方向由操作者自行选择或按打印机缺省设置；1：纵向打印,固定纸张；2：横向打印，固定纸张；3：纵向打印，宽度固定，高度按打印内容的高度自适应。），纸张宽(mm)，纸张高(mm),纸张名(必须纸张宽等于零时本参数才有效。))
                                LODOP.ADD_PRINT_HTM('1%', '1%', '98%', '98%', strFormHtml)        // 设置打印内容
                                // LODOP.ADD_PRINT_TEXT('1%', '1%', '98%', '98%', strFormHtml)    // 设置打印内容
//                                LODOP.ADD_PRINT_BARCODE( 85, 55, 230, 60, '128Auto', item.code);   // 条码（六个参数：Top,Left,Width,Height,BarCodeType,BarCodeValue）
                                LODOP.ADD_PRINT_BARCODE( 10, 5, 260, 60, 'QRCode', printItem.SNCode)   // 条码（六个参数：Top,Left,Width,Height,BarCodeType,BarCodeValue）
                                // LODOP.SET_PREVIEW_WINDOW(2, 0, 0, 800, 600, '')  // 设置预览窗口模式和大小
                                //LODOP.PREVIEW()  // 预览。（这里可以进行预览，注意这里打开时记得把下面的print先注释。）另外，这里的预览只显示单个商品 打印出来的效果即该预览效果。
                                LODOP.PRINT();　　// 打印
                }
        },
    /** 查询装箱列表 */
    getList() {
      this.loading = true
      listPost(this.queryParams).then(response => {
        this.boxRelationList = response.data.list
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
    },
    parseQRCode(code) {

      // var sn={SNCode:code,WorkCode:'111',BatchCode:'dooc',createdAt:''}
      // this.SNList.unshift(sn)
      // this.printItem=sn
      // this.reload()
      this.queryParams.snCode=code
      listPost(this.queryParams).then(response => {
        this.total = response.data.count
        this.loading = false
        if(this.total===1){
          this.SNList.unshift(response.data.list[0])
          this.printItem=response.data.list[0]
         
          this.doPrint(response.data.list[0])
        }
        else{
          this.warningAudio()
        }
      })
    }
    
  },
  mounted() {
          //this.timer = setInterval(this.reload, 10000);
  }
}
</script>


<style lang='scss' scoped>
    .dayinID{width: 237px;height: 155px;border: 1px solid #000;margin-top: 10px;}
    .row5,.row1{ width: 100%;height: 20px;line-height: 20px;color: #000;text-align: center;font-weight: 700;font-size: 0.8rem;}
    .row2{ width: 100%;height: 25px;text-align: center;color: #000;font-weight: 700;font-size: 1.2rem;line-height: 25px;}
    .row3{ width: 100%;height: 20px;line-height: 20px;text-align: center;font-size: 0.8rem;}
    .row4{ width: 100%;height: 60px;}
    .tiaoma-space{width: 100%;height: 10px;margin-top: 20px;}

    .table, .table * {margin: 0 auto; padding: 0;font-size: 14px;font-family: Arial, 宋体, Helvetica, sans-serif;}   
    .table {display: table; width: 200px; border-collapse: collapse;}   
    .table-tr {display: table-row; height: 30px;}   
    .table-th {display: table-cell;font-weight: bold;height: 100%;border: 1px solid gray;text-align: center;vertical-align: middle;}   
    .table-td {display: table-cell; height: 100%;border: 1px solid gray; text-align: center;vertical-align: middle;}  
</style>