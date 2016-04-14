package main.com.iotracks.core_networking.local_client.public_client;

import io.netty.channel.Channel;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;

/**
 * local container connection handler
 *
 * Created by saeid on 4/12/16.
 */
public class PublicLocalClientHandler extends SimpleChannelInboundHandler<byte[]> {
    private final Channel comSatChannel;

    public PublicLocalClientHandler(Channel comSatChannel) {
        this.comSatChannel = comSatChannel;
    }

    /**
     * disconnects from ComSat server when connection to local container been lost.
     *
     * @param ctx
     * @throws Exception
     */
    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        if (comSatChannel != null)
            comSatChannel.disconnect();
    }

    /**
     * pipes received data from local container to ComSat server
     *
     * @param ctx
     * @param bytes
     * @throws Exception
     */
    @Override
    protected void channelRead0(ChannelHandlerContext ctx, byte[] bytes) throws Exception {
        try {
            comSatChannel.writeAndFlush(bytes).sync();
        } catch (Exception e) {}
    }
}
