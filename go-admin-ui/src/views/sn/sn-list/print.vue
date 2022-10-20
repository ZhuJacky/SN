<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">
        <el-form ref="printConf" :model="printConf" :inline="true" label-width="120px">
            <el-form-item label="是否打印UDI" prop="HasUDI">
              <el-radio-group  v-model="printConf.HasUDI">
                <el-radio
                  :key="0"
                  :label="0"
                >否</el-radio>
                <el-radio
                  :key="1"
                  :label="1"
                >是</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="打印数量" prop="Number">
              <el-input-number :disabled="noEdit" readonly="noEdit" v-model="printConf.Number" controls-position="right" :min="0" />
            </el-form-item>
            <el-form-item label="列数" prop="ColNum">
            <el-select v-model="printConf.ColNum" placeholder="打印列数" clearable size="small">
              <el-option
                v-for="dict in ColList"
                :key="dict.value"
                :label="dict.label"
                :value="dict.value"
              />
            </el-select>
          </el-form-item>

        </el-form>
        <el-row :gutter="10" class="mb8">
          <el-col :span="1.5">
            <input  v-model="codeValue" placeholder="请输入条形码"/>
          </el-col>
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
import { listPost, getPost, updatePost,print } from '@/api/sn/sn-info'
import { formatJson } from '@/utils'
import moment from 'moment'
import {clean,printparamsJsonArray,PTK_OpenUSBPort, PTK_CloseUSBPort, PTK_ClearBuffer,PTK_SetDarkness,PTK_SetPrintSpeed,PTK_SetDirection,PTK_SetLabelHeight,PTK_SetLabelWidth,PTK_DrawText_TrueType,PTK_PrintLabel,PTK_DrawBar2D_DATAMATRIX} from '@/utils/POSTEK'
import request from '@/utils/request'

export default {
  name: 'BoxRelationInfoManage',
  inject:['reload'], 
  data() {
    return {
              timer: '',
      goodCodeList:[
        　　　　 　　{uid: 1, brandName: '苹果',goodName: 'iphone13 Pro', code: '11112222333',price: 900, unit:'个', skuName: '8+526G', num: 1},
        ],

    ColList:[{value:1,label:'第一列'},{value:2,label:'第二列'},{value:3,label:'第三列'}],
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
      printConf:{HasUDI:1,Number:1,ColNum:1},
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
    doPrint(printItem) {
//        alert(this.printConf.HasUDI+'aaa'+this.printConf.ColNum)
      	PTK_OpenUSBPort(255)//打开打印机USB端口
	      this.PrintContent(printItem)         //打印内容
		    PTK_CloseUSBPort()  //关闭USB端口
		    this.printing()          //请求数据并打印
    },
    PrintContent(printItem){
	 var mm=12;
		 var column=3; //标签有多少列
		 var printNum=this.printConf.Number; //打印份数
		 var width=34*mm;//单列标签的宽度 （每张小标签的宽度）
		 PTK_ClearBuffer();     //*清空缓存
		 PTK_SetDarkness(20);   //设置打印黑度 取值范围 0-20
		 PTK_SetPrintSpeed(8);  //设置打印速度
		 PTK_SetDirection('B'); //设置打印方向
		 PTK_SetLabelHeight(27*mm,24,24,false); //*设置标签高度、间隙及偏移
		 PTK_SetLabelWidth(104*mm);//*设置标签宽度(底纸的宽度)，一定要准确，否则会导致打印内容位置不准确
  
    for (var i = 1; i < printNum+1; i++) {
			var row=i; //计算荡当前处于第几行
			var col=this.printConf.ColNum;
			var row_cr=col==0?row:row+1;  //如果取余得到的行号 为0则为
			var col_cr=col==0?column:col; //当前处于的列数
		  //PTK_DrawBar2D_QREx((col_cr-1)*width+10,10,0,5,1,0,8,"ss","博思得科技发展有限公司");//打印一个QR码（二维码） 				可根据版本号来固定二维码大小（注意：数据量超过版本所能容纳的量则打印失败） 版本号为0则根据内容自动生成二维码（大小固定）
      PTK_DrawBar2D_DATAMATRIX((col_cr-1)*width+60,70,0,0,0,8,printItem.SNCode)
      if(this.printConf.HasUDI===1){
        PTK_DrawText_TrueType((col_cr-1)*width+35,210,3*mm,0,"Arial",1,900,0,0,0,printItem.UDI); //打udi
      }
      PTK_DrawText_TrueType((col_cr-1)*width+35,245,3*mm,0,"Arial",1,900,0,0,0,printItem.BatchCode); //打印
      PTK_DrawText_TrueType((col_cr-1)*width+35,280,3*mm,0,"Arial",1,900,0,0,0,printItem.SNCode); //打
      
			console.log("-------打印第"+row_cr+"行，第"+col_cr+"列");
			PTK_PrintLabel(1,1); //打印，必须要执行这一行，否则不会打印
		  } 
	},
 printing(){
	    	var data = {};
	    	data.reqParam = "1";
	      	data.printparams = JSON.stringify(printparamsJsonArray);
	     	//jQuery.support.cors = true;  //适用于IE浏览器
			clean(); //此函数必须使用
      console.log(data.printparams)
      var url = "http://127.0.0.1:888/postek/print";
      //return request({    url: url,method: 'post',data: data,dataType:"json"});
      print(url,data).then(response => {
        
        if (response.retval == '0') {
	      
	        		} else {
	        			alert("存在错误，返回结果："+response.msg);
	        		}
      })
      //console.log(JSON.stringify(res));
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
      this.queryParams.mixQRCode=code
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