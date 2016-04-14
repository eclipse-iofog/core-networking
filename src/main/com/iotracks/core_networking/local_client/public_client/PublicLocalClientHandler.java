package main.com.iotracks.core_networking.local_client.public_client;

import io.netty.channel.Channel;
import io.netty.channel.ChannelFuture;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInboundHandlerAdapter;

/**
 * Created by saeid on 4/12/16.
 */
public class PublicLocalClientHandler extends ChannelInboundHandlerAdapter {
    private final Channel comSatChannel;

    public PublicLocalClientHandler(Channel comSatChannel) {
        this.comSatChannel = comSatChannel;
    }

    @Override
    public void channelRead(ChannelHandlerContext ctx, Object msg) throws Exception {
        ChannelFuture future = comSatChannel.writeAndFlush(msg);
        try {
            future.sync();
        } catch (Exception e) {
        }
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        if (comSatChannel != null)
            comSatChannel.disconnect();
    }
}
