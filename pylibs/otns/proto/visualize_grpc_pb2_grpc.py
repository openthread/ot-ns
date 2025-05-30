# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

from . import visualize_grpc_pb2 as visualize__grpc__pb2


class VisualizeGrpcServiceStub(object):
  # missing associated documentation comment in .proto file
  pass

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.Visualize = channel.unary_stream(
        '/visualize_grpc_pb.VisualizeGrpcService/Visualize',
        request_serializer=visualize__grpc__pb2.VisualizeRequest.SerializeToString,
        response_deserializer=visualize__grpc__pb2.VisualizeEvent.FromString,
        )
    self.Command = channel.unary_unary(
        '/visualize_grpc_pb.VisualizeGrpcService/Command',
        request_serializer=visualize__grpc__pb2.CommandRequest.SerializeToString,
        response_deserializer=visualize__grpc__pb2.CommandResponse.FromString,
        )
    self.Energy = channel.unary_stream(
        '/visualize_grpc_pb.VisualizeGrpcService/Energy',
        request_serializer=visualize__grpc__pb2.EnergyRequest.SerializeToString,
        response_deserializer=visualize__grpc__pb2.EnergyEvent.FromString,
        )
    self.NodeStats = channel.unary_stream(
        '/visualize_grpc_pb.VisualizeGrpcService/NodeStats',
        request_serializer=visualize__grpc__pb2.NodeStatsRequest.SerializeToString,
        response_deserializer=visualize__grpc__pb2.VisualizeEvent.FromString,
        )


class VisualizeGrpcServiceServicer(object):
  # missing associated documentation comment in .proto file
  pass

  def Visualize(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def Command(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def Energy(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def NodeStats(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_VisualizeGrpcServiceServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'Visualize': grpc.unary_stream_rpc_method_handler(
          servicer.Visualize,
          request_deserializer=visualize__grpc__pb2.VisualizeRequest.FromString,
          response_serializer=visualize__grpc__pb2.VisualizeEvent.SerializeToString,
      ),
      'Command': grpc.unary_unary_rpc_method_handler(
          servicer.Command,
          request_deserializer=visualize__grpc__pb2.CommandRequest.FromString,
          response_serializer=visualize__grpc__pb2.CommandResponse.SerializeToString,
      ),
      'Energy': grpc.unary_stream_rpc_method_handler(
          servicer.Energy,
          request_deserializer=visualize__grpc__pb2.EnergyRequest.FromString,
          response_serializer=visualize__grpc__pb2.EnergyEvent.SerializeToString,
      ),
      'NodeStats': grpc.unary_stream_rpc_method_handler(
          servicer.NodeStats,
          request_deserializer=visualize__grpc__pb2.NodeStatsRequest.FromString,
          response_serializer=visualize__grpc__pb2.VisualizeEvent.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'visualize_grpc_pb.VisualizeGrpcService', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))
